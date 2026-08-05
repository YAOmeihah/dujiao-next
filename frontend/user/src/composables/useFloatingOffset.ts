export interface FloatingActionBottomOffsetInput {
  defaultBottom: number
  anchorHeight: number
  anchorBottom: number
  gap?: number
}

export const resolveFloatingActionBottomOffset = ({
  defaultBottom,
  anchorHeight,
  anchorBottom,
  gap = 16,
}: FloatingActionBottomOffsetInput) =>
  Math.max(defaultBottom, Math.max(anchorHeight, 0) + Math.max(anchorBottom, 0) + gap)
