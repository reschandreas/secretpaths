<script lang="ts">
	import type { GraphEntry } from '../../types';
	import SubwayChart from './SubwayChart.svelte';

	/** @type {import('../$types').PageData} */
	export let data;

	let newData: GraphEntry = data.graph;

	interface SubwayStation {
		id: string;
		name: string;
		level: number,
		parents?: string[];
	}

	const mapGraphEntryToSubwayStation = (entry: GraphEntry): SubwayStation[] => {
		const children = entry.children;
		if (!children) {
			return [];
		}
		const stations: SubwayStation[] = [];
		children.forEach((child) => {
			const station: SubwayStation = {
				id: child.id,
				name: child.name,
				level: child.level,
				parents: [entry.id]
			};
			stations.push(station);
			if (child.children) {
				const childStations = mapGraphEntryToSubwayStation(child);
				if (childStations) {
					stations.push(...childStations);
				}
			}
		});
		return stations;
	};

	const singleEntry = mapGraphEntryToSubwayStation(newData);

	const root: SubwayStation = { id: '/', name: '/', level: 0, parents: [] };
	singleEntry.push(root);

	const withLevels = singleEntry.reduce((acc: SubwayStation[][], station) => {
		if (!acc[station.level]) {
			acc[station.level] = [];
		}
		acc[station.level].push(station);
		return acc;
	}, []);

	const paths: SubwayStation[][] = withLevels;
</script>
<div class="p-5 w-full h-full">
	<div class="container h-full w-full mx-auto flex-col justify-center">
		<div class="flex flex-col">
			<h2 class="h2 mt-2 font-thin">check out your secretpaths</h2>
		</div>
		<SubwayChart data={paths} />
	</div>
</div>