import Image from "next/image";

export default function Home() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen">
      <h1 className="text-3xl font-bold mb-8">Bienvenido al Sistema de Autenticación</h1>
      <a 
        href="/auth" 
        className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        Ir a la página de autenticación
      </a>
    </div>
  );
}
