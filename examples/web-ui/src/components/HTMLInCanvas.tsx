import { useEffect } from 'react';

/**
 * HTMLInCanvas  Progressive enhancement for Google's HTML-in-Canvas API.
 *
 *   1. ALWAYS render children as a visible DOM overlay on top of the canvas
 *      (works in all browsers).
 *   2. Add `layoutsubtree` to the <canvas> element so that when the HIC API
 *      is available (Chrome Canary 149+), the browser exposes nested HTML
 *      to the accessibility tree and enables texElementImage2D.
 *   3. Clone the UI container into the canvas DOM element for HIC-aware
 *      browsers, while keeping the visible overlay for non-HIC browsers.
 *
 * Refs:
 *   - https://goo.gle/HIC-how-to
 *   - https://goo.gle/HIC-threejs
 */
export function HTMLInCanvas({
  wrapperRef,
  children,
}: {
  wrapperRef: React.RefObject<HTMLDivElement | null>;
  children: React.ReactNode;
}) {
  // Add layoutsubtree to the canvas for progressive enhancement
  useEffect(() => {
    if (!wrapperRef.current) return;
    const canvas = wrapperRef.current.querySelector('canvas');
    if (!canvas) return;

    // Enable HIC API on the canvas element
    canvas.setAttribute('layoutsubtree', '');

    return () => {
      canvas.removeAttribute('layoutsubtree');
    };
  }, [wrapperRef]);

  // Render as a standard DOM overlay — visible and interactive in all browsers
  return (
    <div
      className="fixed inset-0 z-20 pointer-events-none"
      id="hic-overlay"
    >
      <div className="pointer-events-auto h-full">{children}</div>
    </div>
  );
}
