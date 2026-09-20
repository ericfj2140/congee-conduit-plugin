#!/usr/bin/env node
/**
 * Docker e2e: install Conduit from GitHub into a fresh Congee, rank with MiniLM, uninstall.
 * Do not set CONDUIT_EMBEDDER=fake.
 */
import { randomBytes } from 'node:crypto'
import { finalizeEvent, generateSecretKey, getPublicKey } from 'nostr-tools/pure'
import WebSocket from 'ws'
import fs from 'node:fs'

const ADMIN = process.env.CONGEE_ADMIN || 'http://congee:3335'
const RELAY = process.env.CONGEE_RELAY || 'ws://congee:3334/'
const PASS = process.env.ADMIN_PASSWORD || 'e2e-admin'
const REPO = process.env.PLUGIN_REPO || 'michmich112/congee-conduit-plugin'
const TAG = process.env.PLUGIN_TAG || ''
const DATA = '/data'

function headers() {
	return { Authorization: `Bearer ${PASS}`, 'Content-Type': 'application/json' }
}

async function wait(ms) {
	await new Promise((r) => setTimeout(r, ms))
}

async function waitHealth(timeoutMs = 120000) {
	const t0 = Date.now()
	const url = RELAY.replace('ws://', 'http://').replace('wss://', 'https://').replace(/\/$/, '') + '/health'
	while (Date.now() - t0 < timeoutMs) {
		try {
			const res = await fetch(url)
			if (res.ok) return
		} catch {
			/* retry */
		}
		await wait(1000)
	}
	throw new Error('congee health timeout')
}

async function githubAsset() {
	if (process.env.PLUGIN_INSTALL_URL && process.env.PLUGIN_INSTALL_SHA256) {
		return { url: process.env.PLUGIN_INSTALL_URL, sha256: process.env.PLUGIN_INSTALL_SHA256 }
	}
	const arch = process.arch === 'arm64' ? 'arm64' : 'amd64'
	const name = `conduit-plugin-linux-${arch}.tar.gz`
	const api = TAG
		? `https://api.github.com/repos/${REPO}/releases/tags/${TAG}`
		: `https://api.github.com/repos/${REPO}/releases/latest`
	const t0 = Date.now()
	while (Date.now() - t0 < 15 * 60 * 1000) {
		const res = await fetch(api, { headers: { 'User-Agent': 'conduit-e2e' } })
		if (res.ok) {
			const rel = await res.json()
			const tar = (rel.assets || []).find((a) => a.name === name)
			const sha = (rel.assets || []).find((a) => a.name === name + '.sha256')
			if (tar?.browser_download_url) {
				let sha256 = ''
				if (sha?.browser_download_url) {
					const txt = await (await fetch(sha.browser_download_url)).text()
					sha256 = txt.trim().split(/\s+/)[0]
				}
				if (!sha256) throw new Error('missing sha256 asset for ' + name)
				return { url: tar.browser_download_url, sha256 }
			}
		}
		console.log('waiting for GitHub Release asset', name)
		await wait(15000)
	}
	throw new Error('release asset not found: ' + name)
}

async function adminJSON(path, opts = {}) {
	const res = await fetch(ADMIN + path, { ...opts, headers: { ...headers(), ...(opts.headers || {}) } })
	const text = await res.text()
	let body = null
	try {
		body = text ? JSON.parse(text) : null
	} catch {
		body = { raw: text }
	}
	if (!res.ok) throw new Error(`${path} ${res.status}: ${text}`)
	return body
}

async function waitReady() {
	const t0 = Date.now()
	while (Date.now() - t0 < 20 * 60 * 1000) {
		const st = await adminJSON('/api/plugins/conduit/status')
		const emb = st.status?.embedder || {}
		const assets = st.status?.assets || {}
		console.log('status', st.ready, emb.source, emb.vector_ranking, assets.model_ok, assets.runtime_ok, emb.error || '')
		if (emb.source === 'explicit_fake') throw new Error('refusing fake embedder')
		if (st.ready && emb.source === 'onnx' && emb.vector_ranking && assets.model_ok && assets.runtime_ok) {
			return st
		}
		await wait(5000)
	}
	throw new Error('plugin not ready with onnx')
}

function connect(url) {
	return new Promise((resolve, reject) => {
		const ws = new WebSocket(url)
		const q = []
		let pending = null
		ws.on('message', (d) => {
			const msg = JSON.parse(d.toString())
			if (pending) {
				pending(msg)
				pending = null
			} else q.push(msg)
		})
		ws.on('error', reject)
		ws.on('open', () => {
			resolve({
				send: (o) => ws.send(JSON.stringify(o)),
				next: () =>
					new Promise((res) => {
						if (q.length) res(q.shift())
						else pending = res
					}),
				close: () => ws.close()
			})
		})
	})
}

async function publish(c, sk, partial) {
	const ev = finalizeEvent(
		{
			kind: partial.kind,
			created_at: Math.floor(Date.now() / 1000),
			tags: partial.tags || [],
			content: partial.content || ''
		},
		sk
	)
	c.send(['EVENT', ev])
	const msg = await c.next()
	if (msg[0] !== 'OK' || msg[2] !== true) throw new Error('EVENT not OK ' + JSON.stringify(msg))
	return ev
}

async function req(c, sub, filter) {
	c.send(['REQ', sub, filter])
	const events = []
	for (let i = 0; i < 50; i++) {
		const m = await c.next()
		if (m[0] === 'EVENT') events.push(m[2])
		if (m[0] === 'EOSE') break
		if (m[0] === 'CLOSED') throw new Error('CLOSED ' + JSON.stringify(m))
	}
	c.send(['CLOSE', sub])
	return events
}

async function waitEmbeddings(min) {
	const t0 = Date.now()
	while (Date.now() - t0 < 120000) {
		const st = await adminJSON('/api/plugins/conduit/status')
		const n = st.status?.embeddings ?? 0
		const mismatch = st.status?.embedding_mismatch ?? 0
		console.log('embeddings', n, 'mismatch', mismatch)
		if (n >= min && mismatch === 0) return
		await wait(2000)
	}
	throw new Error('embeddings not stored')
}

function assert(cond, msg) {
	if (!cond) throw new Error(msg)
}

async function main() {
	console.log('wait health')
	await waitHealth()
	const asset = await githubAsset()
	console.log('install', asset.url)
	await adminJSON('/api/plugins/install', {
		method: 'POST',
		body: JSON.stringify({ url: asset.url, sha256: asset.sha256, enable: true })
	})
	await waitReady()
	const sk = generateSecretKey()
	const c = await connect(RELAY)
	const bike = await publish(c, sk, {
		kind: 30402,
		content: 'A lightweight mountain bicycle for forest trails and gravel paths.',
		tags: [
			['d', 'bike-' + randomBytes(4).toString('hex')],
			['title', 'Trail Bicycle']
		]
	})
	const pizza = await publish(c, sk, {
		kind: 30402,
		content: 'Wood-fired sourdough pizza with basil and fresh mozzarella.',
		tags: [
			['d', 'food-' + randomBytes(4).toString('hex')],
			['title', 'Napoli Pizza']
		]
	})
	await waitEmbeddings(2)
	const hits = await req(c, 'vec', { kinds: [30402], search: 'two-wheel cycling on dirt tracks' })
	console.log(
		'rank',
		hits.map((e) => e.id),
		'bike',
		bike.id,
		'pizza',
		pizza.id
	)
	assert(hits[0] && hits[0].id === bike.id, 'expected bicycle first for paraphrased search (MiniLM, not bag-of-words)')
	c.close()
	await adminJSON('/api/plugins/conduit/uninstall', {
		method: 'POST',
		body: JSON.stringify({ wipe_data: true })
	})
	const list = await adminJSON('/api/plugins')
	const still = (list.plugins || []).some((p) => p.id === 'conduit')
	assert(!still, 'conduit still listed after uninstall')
	assert(!fs.existsSync(DATA + '/plugins/conduit'), 'leftover plugins/conduit after wipe')
	console.log('PASS conduit docker e2e')
}

main().catch((e) => {
	console.error(e)
	process.exit(1)
})
