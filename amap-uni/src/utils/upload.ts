import type { UploaderFileListItem } from 'vant'

export function fileFromUploader(item: UploaderFileListItem): File | null {
  if (item.file instanceof File) return item.file
  return null
}

export { getActiveDriverTrip, setActiveDriverTrip, normalizeGrabOrder, isGrabSuccess } from './driverTrip'
