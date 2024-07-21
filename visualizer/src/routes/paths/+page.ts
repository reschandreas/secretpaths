import { error } from '@sveltejs/kit';

/** @type {import('../$types').Path[]} */
export const load = async ({ fetch }) => {
	return await fetch('http://localhost:8080/analyzedSecrets')
		.then(response => response.json())
		.then(analyzedSecrets => {
			return {
				analyzedSecrets: analyzedSecrets
			};
		})
		.catch(err => {
			console.error(err);
			return error(500, 'Failed to fetch paths');
		});
};