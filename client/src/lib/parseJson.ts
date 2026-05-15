export type JsonParseError = {
	message: string;
	line: number | null;
	column: number | null;
};

export type JsonParseResult =
	| { ok: true; value: unknown }
	| { ok: false; error: JsonParseError };

/** Parses JSON and extracts line/column from SyntaxError when available. */
export function parseJsonStrict(text: string): JsonParseResult {
	try {
		return { ok: true, value: JSON.parse(text) };
	} catch (e) {
		if (!(e instanceof SyntaxError)) {
			return {
				ok: false,
				error: { message: String(e), line: null, column: null },
			};
		}

		const msg = e.message;
		let line: number | null = null;
		let column: number | null = null;

		const lineMatch = msg.match(/line (\d+)/i);
		const colMatch = msg.match(/column (\d+)/i);
		if (lineMatch) line = Number.parseInt(lineMatch[1], 10);
		if (colMatch) column = Number.parseInt(colMatch[1], 10);

		if (line === null) {
			const posMatch = msg.match(/position (\d+)/i);
			if (posMatch) {
				const pos = Number.parseInt(posMatch[1], 10);
				const before = text.slice(0, pos);
				line = (before.match(/\n/g) || []).length + 1;
				const lastNewline = before.lastIndexOf('\n');
				column = pos - lastNewline;
			}
		}

		const cleanMessage = msg
			.replace(/\s+at position \d+.*$/i, '')
			.replace(/\s*\(line \d+ column \d+\)$/i, '')
			.trim();

		return {
			ok: false,
			error: { message: cleanMessage || msg, line, column },
		};
	}
}
