// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
// and what to do when importing types
import type { AnalyzedSecret, GraphEntry, Information, Path, Policy } from './types';

declare namespace App {
	// interface Locals {}
	interface PageData {
		policies: Policy[];
		paths: Path[];
		analyzedSecrets: AnalyzedSecret[];
		graph: GraphEntry;
		information: Information;
	}
	// interface Error {}
	// interface Platform {}
}
