<script lang="ts">
	import type { CompressedGraphEntry, Information } from '../../types';
	import SubwayChart from './SubwayChart.svelte';
	import { onMount } from 'svelte';

	/** @type {import('../$types').PageData} */
	export let data;

	let newData: CompressedGraphEntry = data.graph;

	let information: Information;

	async function loadInformation() {
		try {
			const res = await fetch(`/v1/info`);
			information = await res.json();
		} catch (err) {
			console.error(err);
		}
	}

	onMount(() => {
		loadInformation();
	});

	interface SubwayStation {
		id: string;
		name: string;
		level: number;
		parent?: string;
	}

	const compressedToSubway = (entry: CompressedGraphEntry, path: string): SubwayStation[] => {
		const children = entry.children;
		if (!children) {
			return [];
		}
		const stations: SubwayStation[] = [];
		children.forEach((child) => {
			// console.log(child);
			let wholePath = "/" + child.prefix;
			if (path !== "/") {
				wholePath = path + wholePath;
			}
			stations.push({
				id: wholePath,
				name: child.prefix,
				level: wholePath.split('/').length - 1,
				parent: path
			});
			stations.push(...compressedToSubway(child, wholePath));
		});
		return stations;
	};

	const singleEntry = compressedToSubway(newData, newData.prefix);

	const root: SubwayStation = { id: '/', name: '/', level: 0, parent: '' };
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
		<SubwayChart data={paths} {information} />
	</div>
</div>
