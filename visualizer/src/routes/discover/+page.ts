import { error } from '@sveltejs/kit';

/** @type {import('../$types').Path[]} */
export const load = async ({ fetch }) => {
	return await fetch('/v1/graph')
		.then((response) => response.json())
		.then((graph) => {
			return {
				graph: graph
			};
		})
		.catch((err) => {
			console.error(err);
			return error(500, 'Failed to fetch graph');
		});
};
