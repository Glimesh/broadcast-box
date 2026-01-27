export default function toBase64Utf8(input: string | undefined): string {
	const utf8Bytes = new TextEncoder().encode(input ?? "")
	const binary = Array.from(utf8Bytes).map(b => String.fromCharCode(b)).join('')

	return btoa(binary)
}
