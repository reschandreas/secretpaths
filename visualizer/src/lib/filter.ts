import type { AnnotatedSecret } from '../types';

export function filterSecrets(secrets: AnnotatedSecret[], query: string): AnnotatedSecret[] {
	if (!query) return secrets;
	const q = query.toLowerCase();
	return secrets.filter((s) => s.path.path.toLowerCase().includes(q));
}
