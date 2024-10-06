import { error } from '@sveltejs/kit';

/** @type {import('../$types').Path[]} */
export const load = async ({ fetch }) => {
	return await fetch('/v1/annotatedSecrets')
		.then((response) => response.json())
		.then((annotatedSecrets) => {
			return {
				annotatedSecrets: annotatedSecrets
			};
		})
		.catch((err) => {
			console.error(err);
			return error(500, 'Failed to fetch paths');
		});
};
