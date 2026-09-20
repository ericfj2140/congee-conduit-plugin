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
});
