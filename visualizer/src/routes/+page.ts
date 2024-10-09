
/** @type {import('./$types').Policy[]} */
export const load = async ({ fetch }) => {
	return await fetch('/v1/policies')
		.then((response) => response.json())
		.then((policies) => {
			return {
				policies: policies
			};
		})
		.catch((err) => {
			console.error(err);
			// return error(500, 'Failed to fetch policies');
		});
};
