'use client';

import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { api, Project, Deployment, Environment } from '@/lib/api';
import { Activity, GitBranch, Clock, TrendingUp, Cpu, HardDrive, Network, DollarSign } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export default function ProjectDetailPage() {
    const params = useParams();
    const projectId = params.id as string;

    const { data: project } = useQuery<Project>({
        queryKey: ['project', projectId],
        queryFn: () => api.getProject(projectId),
    });

    const { data: deployments } = useQuery<Deployment[]>({
        queryKey: ['deployments', projectId],
        queryFn: () => api.getDeployments(projectId),
    });

    const { data: environments } = useQuery<Environment[]>({
        queryKey: ['environments', projectId],
        queryFn: () => api.getEnvironments(projectId),
    });

    // Mock metrics data
    const metricsData = Array.from({ length: 24 }, (_, i) => ({
        time: `${i}:00`,
        cpu: Math.random() * 100,
        memory: Math.random() * 100,
        requests: Math.random() * 1000,
    }));

    if (!project) {
        return <div className="min-h-screen bg-gray-900 flex items-center justify-center">
            <div className="text-white">Loading...</div>
        </div>;
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
            {/* Header */}
            <header className="border-b border-gray-800 bg-gray-900/50 backdrop-blur-xl">
                <div className="max-w-7xl mx-auto px-6 py-6">
                    <div className="flex items-center justify-between">
                        <div>
                            <h1 className="text-3xl font-bold text-white mb-2">{project.name}</h1>
                            <div className="flex items-center gap-4 text-sm text-gray-400">
                                <span>{project.language} • {project.framework}</span>
                                <span>•</span>
                                <div className="flex items-center gap-2">
                                    <GitBranch className="w-4 h-4" />
                                    {project.gitRepo}
                                </div>
                            </div>
                        </div>
                        <button className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition font-medium">
                            Deploy Now
                        </button>
                    </div>
                </div>
            </header>

            <main className="max-w-7xl mx-auto px-6 py-8">
                {/* Environments */}
                <div className="mb-8">
                    <h2 className="text-xl font-bold text-white mb-4">Environments</h2>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {environments?.map((env) => (
                            <div key={env.id} className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                                <div className="flex items-center justify-between mb-4">
                                    <h3 className="text-lg font-semibold text-white capitalize">{env.name}</h3>
                                    <div className={`px-3 py-1 rounded-full text-xs font-medium ${env.status === 'active'
                                            ? 'bg-green-500/20 text-green-400'
                                            : 'bg-gray-500/20 text-gray-400'
                                        }`}>
                                        {env.status}
                                    </div>
                                </div>
                                {env.url && (
                                    <a href={env.url} target="_blank" rel="noopener noreferrer"
                                        className="text-blue-400 hover:text-blue-300 text-sm">
                                        {env.url} →
                                    </a>
                                )}
                            </div>
                        )) || (
                                <div className="col-span-3 text-center py-8 text-gray-400">
                                    No environments configured
                                </div>
                            )}
                    </div>
                </div>

                {/* Metrics */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                    <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                        <div className="flex items-center justify-between mb-4">
                            <h3 className="text-lg font-semibold text-white">CPU Usage</h3>
                            <Cpu className="w-5 h-5 text-blue-400" />
                        </div>
                        <ResponsiveContainer width="100%" height={200}>
                            <LineChart data={metricsData}>
                                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                                <XAxis dataKey="time" stroke="#9CA3AF" />
                                <YAxis stroke="#9CA3AF" />
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
                                    labelStyle={{ color: '#F3F4F6' }}
                                />
                                <Line type="monotone" dataKey="cpu" stroke="#3B82F6" strokeWidth={2} dot={false} />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>

                    <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                        <div className="flex items-center justify-between mb-4">
                            <h3 className="text-lg font-semibold text-white">Memory Usage</h3>
                            <HardDrive className="w-5 h-5 text-purple-400" />
                        </div>
                        <ResponsiveContainer width="100%" height={200}>
                            <LineChart data={metricsData}>
                                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                                <XAxis dataKey="time" stroke="#9CA3AF" />
                                <YAxis stroke="#9CA3AF" />
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
                                    labelStyle={{ color: '#F3F4F6' }}
                                />
                                <Line type="monotone" dataKey="memory" stroke="#A855F7" strokeWidth={2} dot={false} />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {/* Recent Deployments */}
                <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                    <h2 className="text-xl font-bold text-white mb-4">Recent Deployments</h2>
                    <div className="space-y-3">
                        {deployments?.slice(0, 10).map((deployment) => (
                            <div key={deployment.id} className="flex items-center gap-4 p-4 bg-gray-900/50 rounded-lg">
                                <div className={`w-3 h-3 rounded-full ${deployment.status === 'success' ? 'bg-green-500' :
                                        deployment.status === 'failed' ? 'bg-red-500' :
                                            'bg-yellow-500 animate-pulse'
                                    }`}></div>
                                <div className="flex-1">
                                    <div className="flex items-center gap-3">
                                        <span className="text-white font-medium">{deployment.version}</span>
                                        <span className="text-gray-400 text-sm">→</span>
                                        <span className="text-gray-400 text-sm capitalize">{deployment.strategy}</span>
                                    </div>
                                    <div className="text-sm text-gray-500 mt-1">
                                        {deployment.gitCommit.substring(0, 7)} • {deployment.deployedBy}
                                    </div>
                                </div>
                                <div className="text-right">
                                    <div className="text-sm text-gray-400">
                                        <Clock className="w-4 h-4 inline mr-1" />
                                        {new Date(deployment.deployedAt).toLocaleString()}
                                    </div>
                                    {deployment.durationSeconds && (
                                        <div className="text-xs text-gray-500 mt-1">
                                            {deployment.durationSeconds}s
                                        </div>
                                    )}
                                </div>
                            </div>
                        )) || (
                                <div className="text-center py-8 text-gray-400">
                                    No deployments yet
                                </div>
                            )}
                    </div>
                </div>
            </main>
        </div>
    );
}
