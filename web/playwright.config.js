import { defineConfig } from '@playwright/test';

export default defineConfig({
	testDir: './test',
	fullyParallel: false,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 1 : 0,
	use: {
		baseURL: 'http://127.0.0.1:4179',
		trace: 'off'
	},
	webServer: {
		command: 'node test/serve.mjs',
		url: 'http://127.0.0.1:4179/',
		reuseExistingServer: !process.env.CI
	}
});
