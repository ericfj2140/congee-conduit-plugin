import { expect, test } from '@playwright/test';

test('plugin UI loads in an opaque-origin sandbox without calling randomUUID', async ({ page }) => {
	const pageErrors = [];
	page.on('pageerror', (err) => pageErrors.push(err.message));
	page.on('console', (msg) => {
		if (msg.type() === 'error') pageErrors.push(msg.text());
	});
	await page.goto('/');
	const iframe = page.frameLocator('#plugin');
	await expect(iframe.locator('#app')).toBeVisible({ timeout: 15_000 });
	await expect(iframe.getByRole('heading', { name: 'Conduit' })).toBeVisible({ timeout: 15_000 });

	const guest = page.frames().find((f) => f !== page.mainFrame());
	if (!guest) throw new Error('sandbox iframe did not load');
	const probe = await guest.evaluate(() => {
		const out = {
			origin: location.origin,
			typeofUUID: typeof crypto?.randomUUID
		};
		try {
			crypto.randomUUID();
			out.called = true;
		} catch (e) {
			out.threw = e instanceof Error ? e.message : String(e);
		}
		return out;
	});
	test.info().annotations.push({ type: 'crypto-probe', description: JSON.stringify(probe) });

	await page.waitForTimeout(1500);
	await expect(iframe.locator('body')).not.toContainText('crypto.randomUUID is not a function');
	await expect(iframe.getByRole('heading', { name: 'Overview' })).toBeVisible();

	const confirmProbe = await guest.evaluate(() => {
		try {
			return { result: window.confirm('probe'), threw: null };
		} catch (e) {
			return { result: null, threw: e instanceof Error ? e.message : String(e) };
		}
	});
	test.info().annotations.push({ type: 'confirm-probe', description: JSON.stringify(confirmProbe) });

	await iframe.getByRole('button', { name: 'Indexes', exact: true }).click();
	await expect(iframe.locator('main')).toHaveAttribute('data-route', 'indexes');
	await expect(iframe.getByRole('heading', { name: 'Indexes' })).toBeVisible();
	await expect(iframe.getByText('Local model shipped with the plugin')).toBeVisible();

	await iframe.getByRole('button', { name: 'Overview', exact: true }).click();
	await expect(iframe.locator('main')).toHaveAttribute('data-route', 'overview');

	await iframe.getByRole('button', { name: 'Rebuild', exact: true }).click();
	await iframe.getByRole('button', { name: 'Confirm rebuild' }).click();
	await expect
		.poll(async () => page.evaluate(() => window.apiCalls.some((c) => String(c.path).includes('/actions/rebuild'))))
		.toBe(true);

	const fatal = pageErrors.filter((msg) =>
		/goos|goarch|randomUUID|is not defined|is not a function/i.test(msg)
	);
	expect(fatal, `iframe errors: ${pageErrors.join(' | ')}`).toEqual([]);
});
