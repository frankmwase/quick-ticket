import { useRef } from 'react';
import { Canvas } from '@react-three/fiber';
import { GridFloor } from './GridFloor';
import { FloatingParticles } from './FloatingParticles';
import { HologramTicket } from './HologramTicket';
import { ParallaxCamera } from './ParallaxCamera';
import { HTMLInCanvas } from './HTMLInCanvas';

interface SceneProps {
  children?: React.ReactNode;
}

export function Scene({ children }: SceneProps) {
  const wrapperRef = useRef<HTMLDivElement>(null);

  return (
    <div className="fixed inset-0 z-0 bg-[#0a0a0a]" ref={wrapperRef}>
      {/* ── Three.js Scene (pure 3D — no DOM children) ────────────── */}
      <Canvas
        camera={{ position: [0, 1.5, 8], fov: 50 }}
        gl={{ alpha: true, antialias: true }}
      >
        <ParallaxCamera />

        {/* Lighting */}
        <ambientLight intensity={0.15} />
        <pointLight position={[5, 5, 5]} color="#00f0ff" intensity={0.8} />
        <pointLight position={[-5, 3, -5]} color="#ffb000" intensity={0.4} />

        {/* Environment */}
        <GridFloor />
        <FloatingParticles />
        <HologramTicket />

        {/* Fog */}
        <fog attach="fog" args={['#0a0a0a', 4, 25]} />
      </Canvas>

      {/* ── HTML-in-Canvas API ────────────────────────────────────── */}
      {/* Portals children INTO the <canvas> DOM element with        */}
      {/* layoutsubtree for accessibility + browser compositing.     */}
      {/* Lives outside the R3F JSX tree to avoid reconciler errors. */}
      <HTMLInCanvas wrapperRef={wrapperRef}>
        {children}
      </HTMLInCanvas>

      {/* ── CRT Scanline overlay ──────────────────────────────────── */}
      <div className="scanlines fixed inset-0 pointer-events-none z-[100]" />
    </div>
  );
}
