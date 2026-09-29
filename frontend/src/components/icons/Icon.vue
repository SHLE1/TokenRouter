<script lang="ts">
import { defineComponent, h, normalizeClass, ref, type PropType } from 'vue'
import { icons, type IconName } from './registry'
import type { IconControls, IconDefinition } from './types'
import { useIconAnimation } from './useIconAnimation'

type IconSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl'
const SIZE_CLASSES: Record<IconSize, [string, string]> = {
  xs: ['h-3', 'w-3'],
  sm: ['h-4', 'w-4'],
  md: ['h-5', 'w-5'],
  lg: ['h-6', 'w-6'],
  xl: ['h-8', 'w-8']
}

// 图形切换时重新挂载内部节点，根 SVG 继续保留焦点、尺寸和外层状态样式。
const Artwork = (props: {
  definition: IconDefinition
  controls: IconControls
  strokeWidth: number
}) => props.definition.render(props.controls, props.strokeWidth)

export default defineComponent({
  name: 'Icon',
  inheritAttrs: false,
  props: {
    name: { type: String as PropType<IconName>, required: true },
    size: { type: String as PropType<IconSize>, default: 'md' },
    strokeWidth: { type: Number, default: 2 },
    animateOnHover: { type: Boolean, default: true }
  },
  setup(props, { attrs, slots }) {
    const svgRef = ref<SVGSVGElement | null>(null)
    const definition = () => icons[props.name]
    const controls = useIconAnimation(
      svgRef,
      definition,
      () => props.animateOnHover
    )

    return () => {
      const className = normalizeClass(attrs.class)
      const [height, width] = SIZE_CLASSES[props.size]
      const labelled = Boolean(attrs['aria-label'] || attrs['aria-labelledby'])

      return h(
        'svg',
        {
          xmlns: 'http://www.w3.org/2000/svg',
          fill: 'none',
          viewBox: '0 0 24 24',
          stroke: 'currentColor',
          'stroke-width': props.strokeWidth,
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          'aria-hidden': labelled ? undefined : true,
          role: labelled ? 'img' : undefined,
          focusable: 'false',
          ...attrs,
          ref: svgRef,
          'data-animated-icon': definition().name,
          // 调用点的显式尺寸优先，避免 h-4 和默认 h-5 同时争夺尺寸。
          class: [
            'shrink-0',
            !attrs.height &&
              !/(?:^|\s)(?:\S+:)?(?:h-|size-)/.test(className) &&
              height,
            !attrs.width &&
              !/(?:^|\s)(?:\S+:)?(?:w-|size-)/.test(className) &&
              width,
            className
          ]
        },
        [
          h(Artwork, {
            key: props.name,
            definition: definition(),
            controls,
            strokeWidth: props.strokeWidth
          }),
          slots.default?.()
        ]
      )
    }
  }
})
</script>
