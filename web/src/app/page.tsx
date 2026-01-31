import type { Metadata } from 'next';
import Link from 'next/link';
import { Rocket, Zap, Shield, BarChart3, DollarSign, Lock, ArrowRight, Check, Terminal } from 'lucide-react';

export const metadata: Metadata = {
    title: 'OpsAgent - DevOps on Autopilot',
    description: 'Stop wrestling with infrastructure. OpsAgent AI analyzes your code, provisions optimal resources, and deploys your applications—all with zero configuration.',
};

export default function LandingPage() {
    return (
        <div className="min-h-screen bg-black text-white">
            {/* Header */}
            <header className="border-b border-gray-800 bg-black/50 backdrop-blur-xl fixed w-full z-50">
                <div className="max-w-7xl mx-auto px-6 py-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                                <Rocket className="w-5 h-5 text-white" />
                            </div>
                            <span className="text-xl font-bold">OpsAgent</span>
                        </div>
                        <nav className="hidden md:flex items-center gap-8">
                            <Link href="#features" className="text-gray-400 hover:text-white transition">Features</Link>
                            <Link href="/pricing" className="text-gray-400 hover:text-white transition">Pricing</Link>
                            <Link href="/docs" className="text-gray-400 hover:text-white transition">Docs</Link>
                            <Link href="/login" className="text-gray-400 hover:text-white transition">Sign in</Link>
                            <Link href="/signup" className="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition font-medium">
                                Start Free Trial
                            </Link>
                        </nav>
                    </div>
                </div>
            </header>

            {/* Hero */}
            <section className="pt-32 pb-20 px-6">
                <div className="max-w-7xl mx-auto text-center">
                    <div className="inline-flex items-center gap-2 px-4 py-2 bg-blue-500/10 border border-blue-500/20 rounded-full text-blue-400 text-sm mb-8">
                        <Rocket className="w-4 h-4" />
                        7-Day Free Trial • No credit card required
                    </div>

                    <h1 className="text-6xl md:text-7xl font-bold mb-6">
                        DevOps on <span className="bg-gradient-to-r from-blue-500 to-purple-600 bg-clip-text text-transparent">Autopilot</span>
                    </h1>

                    <p className="text-xl text-gray-400 max-w-3xl mx-auto mb-12">
                        Stop wrestling with infrastructure. OpsAgent AI analyzes your code,
                        provisions optimal resources, and deploys your applications—all with zero configuration.
                    </p>

                    {/* Stats */}
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-8 max-w-4xl mx-auto mb-12">
                        <div>
                            <div className="text-4xl font-bold text-blue-500 mb-2">94%</div>
                            <div className="text-sm text-gray-400">Detection Accuracy</div>
                        </div>
                        <div>
                            <div className="text-4xl font-bold text-blue-500 mb-2">2.5s</div>
                            <div className="text-sm text-gray-400">Deploy Time</div>
                        </div>
                        <div>
                            <div className="text-4xl font-bold text-blue-500 mb-2">70%</div>
                            <div className="text-sm text-gray-400">Cost Savings</div>
                        </div>
                        <div>
                            <div className="text-4xl font-bold text-blue-500 mb-2">10+</div>
                            <div className="text-sm text-gray-400">Services Detected</div>
                        </div>
                    </div>

                    {/* Terminal Demo */}
                    <div className="max-w-3xl mx-auto bg-gray-900 rounded-xl border border-gray-800 overflow-hidden mb-12">
                        <div className="flex items-center gap-2 px-4 py-3 bg-gray-800 border-b border-gray-700">
                            <div className="flex gap-2">
                                <div className="w-3 h-3 rounded-full bg-red-500"></div>
                                <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
                                <div className="w-3 h-3 rounded-full bg-green-500"></div>
                            </div>
                            <span className="text-sm text-gray-400 ml-4">Terminal — ops</span>
                        </div>
                        <div className="p-6 font-mono text-sm text-left">
                            <div className="text-green-400">$ npm install -g @opsagent/cli</div>
                            <div className="text-green-400 mt-4">$ ops init</div>
                            <div className="text-gray-400 mt-2">✓ Detected: Node.js + Next.js</div>
                            <div className="text-gray-400">✓ Services: PostgreSQL, Redis, Stripe</div>
                            <div className="text-gray-400">✓ Est. cost: $85/month</div>
                            <div className="text-green-400 mt-4">$ ops deploy</div>
                            <div className="text-gray-400 mt-2">✓ Built in 45s</div>
                            <div className="text-gray-400">✓ Deployed to production</div>
                            <div className="text-blue-400 mt-2">→ https://my-app.opsagent.dev</div>
                        </div>
                    </div>

                    <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                        <Link href="/signup" className="px-8 py-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 rounded-lg font-semibold text-lg transition flex items-center gap-2">
                            Start 7-Day Free Trial
                            <ArrowRight className="w-5 h-5" />
                        </Link>
                        <Link href="/docs" className="px-8 py-4 bg-gray-800 hover:bg-gray-700 rounded-lg font-semibold text-lg transition">
                            View Documentation
                        </Link>
                    </div>
                </div>
            </section>

            {/* Features */}
            <section id="features" className="py-20 px-6 bg-gradient-to-b from-black to-gray-900">
                <div className="max-w-7xl mx-auto">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl md:text-5xl font-bold mb-4">Everything You Need</h2>
                        <p className="text-xl text-gray-400">Enterprise-grade DevOps, zero configuration required</p>
                    </div>

                    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
                        {[
                            {
                                icon: Zap,
                                title: 'AI-Powered Analysis',
                                description: 'Automatically detects language, framework, dependencies, and resource requirements with 94%+ accuracy.',
                                color: 'from-yellow-500 to-orange-500'
                            },
                            {
                                icon: Rocket,
                                title: 'Zero-Config Deploys',
                                description: 'Deploy any app without writing Dockerfiles, YAML, or infrastructure code. Just run ops deploy.',
                                color: 'from-blue-500 to-cyan-500'
                            },
                            {
                                icon: BarChart3,
                                title: 'Smart Deployments',
                                description: 'Blue-green, canary, and rolling deployments with automatic rollback on failures.',
                                color: 'from-purple-500 to-pink-500'
                            },
                            {
                                icon: Shield,
                                title: 'Built-in Monitoring',
                                description: 'Metrics, logs, and 5 alert rules configured automatically for every deployment.',
                                color: 'from-green-500 to-emerald-500'
                            },
                            {
                                icon: DollarSign,
                                title: 'Cost Optimization',
                                description: 'AI-driven spot instance recommendations to reduce cloud spend by 70%+.',
                                color: 'from-green-500 to-teal-500'
                            },
                            {
                                icon: Lock,
                                title: 'Security by Default',
                                description: 'Secrets scanning, vulnerability detection, automatic SSL, and compliance automation.',
                                color: 'from-red-500 to-rose-500'
                            },
                        ].map((feature, index) => (
                            <div key={index} className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-8 hover:border-gray-600 transition group">
                                <div className={`w-12 h-12 bg-gradient-to-br ${feature.color} rounded-lg flex items-center justify-center mb-4 group-hover:scale-110 transition`}>
                                    <feature.icon className="w-6 h-6 text-white" />
                                </div>
                                <h3 className="text-xl font-bold mb-3">{feature.title}</h3>
                                <p className="text-gray-400">{feature.description}</p>
                            </div>
                        ))}
                    </div>
                </div>
            </section>

            {/* How It Works */}
            <section className="py-20 px-6">
                <div className="max-w-7xl mx-auto">
                    <div className="text-center mb-16">
                        <h2 className="text-4xl md:text-5xl font-bold mb-4">Deploy in 3 Simple Steps</h2>
                        <p className="text-xl text-gray-400">From code to production in minutes</p>
                    </div>

                    <div className="grid md:grid-cols-3 gap-8">
                        {[
                            {
                                step: '1',
                                title: 'Install CLI',
                                code: 'npm install -g @opsagent/cli',
                                description: 'One command to get started'
                            },
                            {
                                step: '2',
                                title: 'Initialize',
                                code: 'ops init — AI analyzes your code',
                                description: 'AI detects everything automatically'
                            },
                            {
                                step: '3',
                                title: 'Deploy',
                                code: 'ops deploy — Live in seconds',
                                description: 'Production-ready infrastructure'
                            },
                        ].map((step, index) => (
                            <div key={index} className="relative">
                                <div className="bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20 rounded-xl p-8">
                                    <div className="text-6xl font-bold text-blue-500/20 mb-4">{step.step}</div>
                                    <h3 className="text-2xl font-bold mb-3">{step.title}</h3>
                                    <div className="bg-black/50 rounded-lg p-4 mb-4 font-mono text-sm text-green-400">
                                        {step.code}
                                    </div>
                                    <p className="text-gray-400">{step.description}</p>
                                </div>
                                {index < 2 && (
                                    <div className="hidden md:block absolute top-1/2 -right-4 transform -translate-y-1/2">
                                        <ArrowRight className="w-8 h-8 text-blue-500/30" />
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            </section>

            {/* CTA */}
            <section className="py-20 px-6">
                <div className="max-w-4xl mx-auto text-center">
                    <div className="bg-gradient-to-br from-blue-500/20 to-purple-500/20 border border-blue-500/30 rounded-2xl p-12">
                        <h2 className="text-4xl md:text-5xl font-bold mb-4">
                            Ready to simplify your DevOps?
                        </h2>
                        <p className="text-xl text-gray-300 mb-8">
                            Join 10,000+ developers who deploy with confidence.
                        </p>
                        <Link href="/signup" className="inline-flex items-center gap-2 px-8 py-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 rounded-lg font-semibold text-lg transition">
                            Start Free Trial
                            <ArrowRight className="w-5 h-5" />
                        </Link>
                        <div className="flex items-center justify-center gap-6 mt-6 text-sm text-gray-400">
                            <div className="flex items-center gap-2">
                                <Check className="w-4 h-4 text-green-500" />
                                7-day free trial
                            </div>
                            <div className="flex items-center gap-2">
                                <Check className="w-4 h-4 text-green-500" />
                                No credit card
                            </div>
                            <div className="flex items-center gap-2">
                                <Check className="w-4 h-4 text-green-500" />
                                Cancel anytime
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Footer */}
            <footer className="border-t border-gray-800 py-12 px-6">
                <div className="max-w-7xl mx-auto">
                    <div className="flex flex-col md:flex-row items-center justify-between gap-4">
                        <div className="flex items-center gap-2">
                            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                                <Rocket className="w-5 h-5 text-white" />
                            </div>
                            <span className="text-xl font-bold">OpsAgent</span>
                        </div>
                        <div className="flex items-center gap-8 text-sm text-gray-400">
                            <Link href="/docs" className="hover:text-white transition">Documentation</Link>
                            <Link href="/pricing" className="hover:text-white transition">Pricing</Link>
                            <Link href="/privacy" className="hover:text-white transition">Privacy</Link>
                            <Link href="/terms" className="hover:text-white transition">Terms</Link>
                        </div>
                        <div className="text-sm text-gray-400">
                            © 2024 OpsAgent. All rights reserved.
                        </div>
                    </div>
                </div>
            </footer>
        </div>
    );
}
