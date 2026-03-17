const CN_DIGITS = ['', '一', '二', '三', '四', '五', '六', '七', '八', '九']
const CN_TEN = '十'

/** 数字转中文序数，1->一, 2->二, 11->十一, 21->二十一 */
export function numberToChinese(n: number): string {
  if (n <= 0 || n > 99) return String(n)
  if (n < 10) return CN_DIGITS[n]
  if (n === 10) return CN_TEN
  if (n < 20) return CN_TEN + CN_DIGITS[n - 10]
  const tens = Math.floor(n / 10)
  const ones = n % 10
  return CN_DIGITS[tens] + CN_TEN + (ones > 0 ? CN_DIGITS[ones] : '')
}

/** 移除章节名中的「（第二版）」等版本后缀 */
export function stripEdition(name: string): string {
  return (name || '').replace(/（第二版）|\(第二版\)/g, '').trim()
}

/** 格式化章标题：第X章 名称 */
export function formatChapterTitle(index: number, name: string): string {
  const cn = numberToChinese(index)
  const clean = stripEdition(name || '')
  return `第${cn}章 ${clean}`
}
