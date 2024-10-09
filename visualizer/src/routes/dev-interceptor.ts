import { browser } from '$app/environment';

function intercept() {
	if (browser) {
		const originalFetch = window.fetch;

		// @ts-ignore
		window.fetch = async (input: RequestInfo, init?: RequestInit) => {
			// Log the request URL and options
			console.log('Client-side fetch to:', input);

			if (typeof input === 'string' || input instanceof Request) {
				const modifiedInit = init || {};

				if (input instanceof Request) {
					const url = new URL(input.url);
					if (url.hostname === 'localhost' && url.port === '5173') {
						url.port = '8081';
						input = new Request(url, input);
					}
					if (url.pathname.startsWith('/v1/')) {
						url.pathname = url.pathname.replace('/v1/', '/v1/');
						input = new Request(url, input);
					}
					input = new Request(url, input);
				}
				if (typeof input === 'string') {
					if (input.startsWith('/v1/')) {
						input = input.replace('/v1/', 'http://localhost:8081/v1/');
					}
					input = input.replace('localhost:5173', 'localhost:8081');
				}
				console.log('Client-side fetch modified to:', input);
				// Proceed with the modified fetch
				return originalFetch(input, modifiedInit);
			}

			// Proceed with the original fetch if no modification is needed
			return originalFetch(input, init);
		};
	}
}

intercept();
