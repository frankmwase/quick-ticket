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
  return (
    <div className="fixed inset-0 z-0 bg-[#0a0a0a]">
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

        {/* HTML UI inside Canvas using experimental HIC API */}
        <HTMLInCanvas>{children}</HTMLInCanvas>

        {/* Fog */}
        <fog attach="fog" args={['#0a0a0a', 4, 25]} />
      </Canvas>

      {/* CRT Scanline overlay (post-processing effect, safely on top) */}
      <div className="scanlines fixed inset-0 pointer-events-none z-[100]" />
    </div>
  );
}

