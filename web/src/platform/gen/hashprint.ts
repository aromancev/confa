type colorValues = 0 | 1 | 2

type allowedAlgs = "SHA-256" | "SHA-384" | "SHA-512"

const cells = 7

interface Opts {
  data: string
  size?: number
  bg?: string
  saturation?: number
  lightness?: number
  likeness?: [number, number]
  algorithm?: allowedAlgs
  scale?: number
}

export async function hashprint(opts: Opts): Promise<string> {
  // Set defaults
  opts.size = opts.size || 140
  opts.bg = opts.bg || "#00000000"
  opts.saturation = opts.saturation || 0.7
  opts.lightness = opts.lightness || 0.5
  opts.algorithm = opts.algorithm || "SHA-256"
  opts.scale = opts.scale || 1

  if (!opts.likeness || opts.likeness[0] + opts.likeness[1] > 1) {
    opts.likeness = [0.5, 0.25]
  }

  // Get the hash
  const digest = await digestData(opts.data, opts.algorithm)
  const view = new Uint8Array(digest)

  // Get the grid values
  const a: colorValues[] = []
  // (50% chance it's 1, 25% it's 2, 25% it's 0)
  for (let i = 0; i < cells * 4; i++) {
    if (view[i] <= 256 * opts.likeness[0]) {
      a.push(1)
    } else if (view[i] <= 256 * (opts.likeness[0] + opts.likeness[1])) {
      a.push(2)
    } else {
      a.push(0)
    }
  }

  const active: colorValues[] = [
    a[0],
    a[1],
    a[2],
    a[3],
    a[2],
    a[1],
    a[0],
    a[4],
    a[5],
    a[6],
    a[7],
    a[6],
    a[5],
    a[4],
    a[8],
    a[9],
    a[10],
    a[11],
    a[10],
    a[9],
    a[8],
    a[12],
    a[13],
    a[14],
    a[15],
    a[14],
    a[13],
    a[12],
    a[16],
    a[17],
    a[18],
    a[19],
    a[18],
    a[17],
    a[16],
    a[20],
    a[21],
    a[22],
    a[23],
    a[22],
    a[21],
    a[20],
    a[24],
    a[25],
    a[26],
    a[27],
    a[26],
    a[25],
    a[24],
  ]

  // Get the hue values from the last 4 bytes (2 bytes per hue)
  // byte1 + byte2 -- 256
  //             X -- 360
  const hue1 = ((view[28] + view[29]) * 360) / 256
  const hue2 = ((view[30] + view[31]) * 360) / 256

  // Create canvas
  const canvas = document.createElement("canvas")
  canvas.width = canvas.height = opts.size
  const ctx = canvas.getContext("2d") as CanvasRenderingContext2D

  // Fill background
  ctx.fillStyle = opts.bg
  ctx.strokeStyle = "#00000000"
  ctx.fillRect(0, 0, opts.size, opts.size)

  const l = Math.floor(opts.size / cells)
  const extra = opts.size % cells

  // Translate and scale for padding.
  const translace = (opts.size - opts.size * opts.scale) / 2
  ctx.translate(translace, translace)
  ctx.scale(opts.scale, opts.scale)

  // Draw cells.
  for (let i = 0; i < active.length; i++) {
    if (active[i] === 0) continue

    ctx.fillStyle = `hsl(${active[i] === 1 ? hue1 : hue2}, ${opts.saturation * 100}%, ${opts.lightness * 100}%)`

    const x = i % cells
    const y = Math.floor(i / cells)

    ctx.fillRect(
      x <= 3 ? x * l : x * l + extra,
      y <= 3 ? y * l : y * l + extra,
      x === 3 ? l + extra : l,
      y === 3 ? l + extra : l,
    )
  }

  return canvas.toDataURL()
}

// Compute hash
const digestData = async (d: string, hashFunction: allowedAlgs): Promise<ArrayBuffer> => {
  const encoder = new TextEncoder()
  const data = encoder.encode(d)
  return await crypto.subtle.digest(hashFunction, data)
}
