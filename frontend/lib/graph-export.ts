type ExportNode = {
  id: string
  label: string
  type: string
  x?: number
  y?: number
}

type ExportLink = {
  source: string | ExportNode
  target: string | ExportNode
  label?: string
}

function nodeRef(value: string | ExportNode): ExportNode | null {
  if (typeof value === "string") return null
  return value
}

function escapeXml(value: string): string {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
}

export function downloadBlob(filename: string, blob: Blob) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement("a")
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

export function exportCanvasPNG(
  canvas: HTMLCanvasElement | null,
  filename: string
): boolean {
  if (!canvas) return false
  const link = document.createElement("a")
  link.download = filename
  link.href = canvas.toDataURL("image/png")
  link.click()
  return true
}

export function exportGraphSVG(options: {
  nodes: ExportNode[]
  links: ExportLink[]
  width: number
  height: number
  nodeColor: (node: ExportNode) => string
  filename: string
}): boolean {
  const positioned = options.nodes.filter(
    (node) => typeof node.x === "number" && typeof node.y === "number"
  )
  if (positioned.length === 0) return false

  const nodeById = new Map(positioned.map((node) => [node.id, node]))
  let body = ""

  for (const link of options.links) {
    const source = nodeRef(link.source)
    const target = nodeRef(link.target)
    const sourceNode =
      source ??
      (typeof link.source === "string" ? nodeById.get(link.source) : undefined)
    const targetNode =
      target ??
      (typeof link.target === "string" ? nodeById.get(link.target) : undefined)
    if (
      !sourceNode ||
      !targetNode ||
      sourceNode.x == null ||
      sourceNode.y == null ||
      targetNode.x == null ||
      targetNode.y == null
    ) {
      continue
    }
    body += `<line x1="${sourceNode.x}" y1="${sourceNode.y}" x2="${targetNode.x}" y2="${targetNode.y}" stroke="#94a3b8" stroke-width="1" />`
  }

  for (const node of positioned) {
    const color = options.nodeColor(node)
    body += `<circle cx="${node.x}" cy="${node.y}" r="5" fill="${color}" />`
    body += `<text x="${node.x}" y="${(node.y ?? 0) - 10}" font-size="9" text-anchor="middle" fill="#334155">${escapeXml(node.label)}</text>`
  }

  const svg = `<?xml version="1.0" encoding="UTF-8"?><svg xmlns="http://www.w3.org/2000/svg" width="${options.width}" height="${options.height}" viewBox="0 0 ${options.width} ${options.height}"><rect width="100%" height="100%" fill="#f8fafc"/>${body}</svg>`
  downloadBlob(options.filename, new Blob([svg], { type: "image/svg+xml" }))
  return true
}