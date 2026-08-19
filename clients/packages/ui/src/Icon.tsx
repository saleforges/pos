import iconColor from '@saleforges/assets/brand/icon/icon_color.svg'
import iconMono from '@saleforges/assets/brand/icon/icon_mono.svg'
import iconWhite from '@saleforges/assets/brand/icon/icon_white.svg'

const VARIANTS = { color: iconColor, mono: iconMono, white: iconWhite }
// Icon SVGs use a 1024:976 viewBox; explicit width avoids distortion.
const ASPECT_RATIO = 1024 / 976

export function Icon({
  variant = 'color',
  size = 32,
}: {
  variant?: keyof typeof VARIANTS
  size?: number
}) {
  return (
    <img
      src={VARIANTS[variant]}
      width={size * ASPECT_RATIO}
      height={size}
      alt="SaleForges"
    />
  )
}