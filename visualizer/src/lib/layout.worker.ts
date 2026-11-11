import { computeLayout } from './layout';
import type { SubwayStation } from './layout';

self.onmessage = (e: MessageEvent<SubwayStation[][]>) => {
	const result = computeLayout(e.data);
	self.postMessage(result);
};
