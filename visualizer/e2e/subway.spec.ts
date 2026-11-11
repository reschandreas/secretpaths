import { test, expect } from '@playwright/test';

const graph = {
	prefix: '/',
	children: [
		{
			prefix: 'secret',
			children: [
				{ prefix: 'foo', children: [] },
				{ prefix: 'bar', children: [] }
			]
		},
		{ prefix: 'sys', children: [{ prefix: 'policy', children: [] }] }
	]
};

test('discover renders subway chart canvas', async ({ page }) => {
	await page.route('**/v1/info', (route) =>
		route.fulfill({ json: { version: '0.0.2', vaultAddress: 'http://vault', kvEngine: 'secret' } })
	);
	await page.route('**/v1/graph*', (route) => route.fulfill({ json: graph }));

	await page.goto('/discover');
	const canvas = page.getByLabel('Subway chart of secret paths');
	await expect(canvas).toBeVisible();
	await page.waitForTimeout(500);

	await expect(page).toHaveScreenshot('subway.png', { maxDiffPixelRatio: 0.02 });
});
