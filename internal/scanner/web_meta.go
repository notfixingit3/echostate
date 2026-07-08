package scanner

const metaIntelJS = `(() => {
	const out = {};
	const add = (key, value) => {
		const text = String(value || '').trim();
		if (text) out[key] = text;
	};

	const description = document.querySelector('meta[name="description"]');
	if (description?.content) add('description', description.content);

	const generator = document.querySelector('meta[name="generator"]');
	if (generator?.content) add('generator', generator.content);

	const canonical = document.querySelector('link[rel="canonical"]');
	if (canonical?.href) add('canonical', canonical.href);

	const openGraph = {};
	document.querySelectorAll('meta[property^="og:"]').forEach((el) => {
		const key = el.getAttribute('property');
		const content = el.getAttribute('content');
		if (key && content) openGraph[key] = content.trim();
	});
	if (Object.keys(openGraph).length > 0) {
		out.open_graph = openGraph;
	}

	return out;
})()`
