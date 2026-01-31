export default function Home() {
    return (
        <main className="flex min-h-screen flex-col items-center justify-center p-24 bg-gradient-to-br from-blue-50 to-indigo-100">
            <div className="max-w-2xl w-full space-y-8 text-center">
                <h1 className="text-6xl font-bold text-gray-900">
                    🚀 OpsAgent Demo
                </h1>
                <p className="text-xl text-gray-600">
                    Production-ready Next.js application deployed with OpsAgent
                </p>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-12">
                    <div className="bg-white p-6 rounded-lg shadow-md">
                        <div className="text-3xl mb-2">⚡</div>
                        <h3 className="font-semibold text-lg">Next.js 14</h3>
                        <p className="text-sm text-gray-600">App Router + Server Components</p>
                    </div>

                    <div className="bg-white p-6 rounded-lg shadow-md">
                        <div className="text-3xl mb-2">🗄️</div>
                        <h3 className="font-semibold text-lg">PostgreSQL</h3>
                        <p className="text-sm text-gray-600">Managed database with backups</p>
                    </div>

                    <div className="bg-white p-6 rounded-lg shadow-md">
                        <div className="text-3xl mb-2">⚡</div>
                        <h3 className="font-semibold text-lg">Redis</h3>
                        <p className="text-sm text-gray-600">High-performance caching</p>
                    </div>
                </div>

                <div className="mt-8 space-y-4">
                    <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                        <p className="text-green-800 font-semibold">✓ Deployed with OpsAgent</p>
                        <p className="text-green-600 text-sm">Auto-scaling • Zero-downtime • Monitoring</p>
                    </div>
                </div>

                <div className="mt-8 flex gap-4 justify-center">
                    <a
                        href="/api/health"
                        className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
                    >
                        Health Check
                    </a>
                    <a
                        href="/api/users"
                        className="px-6 py-3 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition"
                    >
                        API Demo
                    </a>
                </div>
            </div>
        </main>
    );
}
