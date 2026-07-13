declare namespace BMapGL {
  class Point {
    constructor(lng: number, lat: number)
    lng: number
    lat: number
  }

  class Map {
    constructor(container: HTMLElement | string)
    centerAndZoom(point: Point, zoom: number): void
    enableScrollWheelZoom(enable: boolean): void
    addOverlay(overlay: Overlay): void
    removeOverlay(overlay: Overlay): void
    setViewport(points: Point[]): void
    addEventListener(event: string, handler: (e: MapClickEvent) => void): void
    destroy(): void
  }

  class Size {
    constructor(width: number, height: number)
  }

  class Icon {
    constructor(url: string, size: Size, options?: { anchor?: Size })
  }

  class Marker {
    constructor(point: Point, options?: { title?: string; icon?: Icon })
  }

  class Polyline {
    constructor(points: Point[], options?: {
      strokeColor?: string
      strokeWeight?: number
      strokeOpacity?: number
    })
  }

  interface Overlay {}

  interface MapClickEvent {
    latlng: { lng: number; lat: number }
  }
}

interface Window {
  BMapGL: typeof BMapGL
}
