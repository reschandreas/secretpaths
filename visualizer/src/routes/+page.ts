import { error } from '@sveltejs/kit';

/** @type {import('./$types').Policy[]} */
export const load = async ({ fetch }) => {
	return await fetch('http://localhost:8081/policies')
		.then(response => response.json())
		.then(policies => {
			return {
				policies: policies
			};
		})
		.catch(err => {
			console.error(err);
			return error(500, 'Failed to fetch policies');
		});
};