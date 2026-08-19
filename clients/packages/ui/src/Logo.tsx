import logoColor from '@saleforges/assets/brand/logo/logo-full-color.svg'
import logoMono from '@saleforges/assets/brand/logo/logo-full-mono.svg'
import logoWhite from '@saleforges/assets/brand/logo/logo-full-white.svg'

const VARIANTS = { color: logoColor, mono: logoMono, white: logoWhite }
// All logo variants share the same viewBox ratio (1151:260). Explicit width keeps the
// image from being distorted by a flex container's default align-items: stretch.
const ASPECT_RATIO = 1151 / 260

export function Logo({
  variant = 'color',
  height = 32,
}: {
  variant?: keyof typeof VARIANTS
  height?: number
}) {
  return (
    <img
      src={VARIANTS[variant]}
      style={{ height, width: height * ASPECT_RATIO, maxWidth: '100%', flexShrink: 0 }}
      alt="SaleForges"
    />
  )
}