import { Canvas } from '@react-three/fiber';
import { GridFloor } from './GridFloor';
import { FloatingParticles } from './FloatingParticles';
import { HologramTicket } from './HologramTicket';
import { ParallaxCamera } from './ParallaxCamera';

interface SceneProps {
  children?: React.ReactNode;
}

export function Scene({ children }: SceneProps) {
  return (
    <div className="fixed inset-0 z-0">
      <Canvas
        camera={{ position: [0, 1.5, 8], fov: 50 }}
        gl={{ alpha: true, antialias: true }}
        style={{ background: '#050505' }}
      >
        <ParallaxCamera />

        {/* Lighting */}
        <ambientLight intensity={0.15} />
        <pointLight position={[5, 5, 5]} color="#00ff41" intensity={0.8} />
        <pointLight position={[-5, 3, -5]} color="#ffb000" intensity={0.4} />

        {/* Environment */}
        <GridFloor />
        <FloatingParticles />
        <HologramTicket />

        {/* Fog */}
        <fog attach="fog" args={['#050505', 8, 30]} />
      </Canvas>

      {/* CRT Scanline overlay */}
      <div className="scanlines fixed inset-0 pointer-events-none z-10" />

      {/* DOM overlay for UI */}
      <div className="fixed inset-0 z-20 pointer-events-none">
        <div className="pointer-events-auto h-full">{children}</div>
      </div>
    </div>
  );
}
