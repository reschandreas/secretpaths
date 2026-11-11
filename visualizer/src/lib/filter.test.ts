import { describe, it, expect } from 'vitest';
import { filterSecrets } from './filter';
import type { AnnotatedSecret } from '../types';

const mk = (p: string): AnnotatedSecret => ({ path: { path: p }, policies: [] });

describe('filterSecrets', () => {
	it('returns all when query empty', () => {
		const s = [mk('secret/foo'), mk('secret/bar')];
		expect(filterSecrets(s, '')).toHaveLength(2);
	});

	it('filters case-insensitively', () => {
		const s = [mk('secret/Foo'), mk('secret/bar')];
		expect(filterSecrets(s, 'FOO')).toHaveLength(1);
		expect(filterSecrets(s, 'foo')[0].path.path).toBe('secret/Foo');
	});

	it('returns empty when no match', () => {
		const s = [mk('secret/foo')];
		expect(filterSecrets(s, 'zzz')).toHaveLength(0);
	});
});
