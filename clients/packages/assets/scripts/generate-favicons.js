// clients/packages/assets/scripts/generate-favicons.js
// Run from packages/assets: `bun run generate-favicons`
import sharp from 'sharp'
import { mkdir } from 'node:fs/promises'
const apps = {
  backoffice: {
    blue: [16, 32],
    dark: [16, 32],
    white: [16, 32, 180],
  },
  intools: {
    blue: [16, 32],
  },
}

for (const [app, variants] of Object.entries(apps)) {
  await mkdir(`../../${app}/public`, { recursive: true }).catch(() => {})
  for (const [variant, sizes] of Object.entries(variants)) {
    for (const size of sizes) {
      await sharp(`brand/app-icon/app-icon-${variant}-bg.svg`)
        .resize(size, size)
        .png()
        .toFile(`../../${app}/public/favicon-${variant}-${size}.png`)
    }
  }
}
