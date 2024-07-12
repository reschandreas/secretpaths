export interface Rule {
	path: string;
	capabilities: string[];
}

export interface Policy {
	name: string;
	rules: Rule[];
}

export interface Path {
	path: string;
}

export interface AnalyzedSecret {
	path: Path;
	policies: Policy[];
}