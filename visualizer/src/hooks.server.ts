import type { HandleFetch } from '@sveltejs/kit';

export const handleFetch: HandleFetch = async ({ request, fetch }) => {
	// Check if the app is in development mode
	if (process.env.NODE_ENV === 'development') {
		// Intercept the request and log the URL
		console.log('Intercepted request to:', request.url);
		request = new Request(
			new URL(request.url.replace('http://localhost:5173', 'http://localhost:8081')),
			request
		);
	}
	return fetch(request);
};
