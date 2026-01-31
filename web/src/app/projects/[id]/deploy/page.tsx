'use client';
import { useParams } from 'next/navigation';
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tantml:react-query';
import { api, Environment } from '@/lib/api';
import { Rocket, GitBranch, Settings, Play } from 'lucide-react';
export default function DeployPage() {
    const params = useParams();
    const projectId = params.id as string;
    const queryClient = useQueryClient();
    const [selectedEnv, setSelectedEnv] = useState('production');
    const [selectedBranch, setSelectedBranch] = useState('main');
    const [selectedStrategy, setSelectedStrategy] = useState('rolling');
    const { data: environments } = useQuery<Environment[]>({
        queryKey: ['environments', projectId],
        queryFn: () => api.getEnvironments(projectId),
    });
    const deployMutation = useMutation({
        mutationFn: () => api.deploy(projectId, {
            environment: selectedEnv,
            strategy: selectedStrategy,
            branch: selectedBranch,
        }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['deployments', projectId] });
        },
    });
    const handleDeploy = () => {
        deployMutation.mutate();
    };
    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
            <div className="max-w-4xl mx-auto px-6 py-12">
                <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-2xl p-8">
                    <div className="flex items-center gap-4 mb-8">
                        <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center">
                            <Rocket className="w-8 h-8 text-white" />
                        </div>
                        <div>
                            <h1 className="text-3xl font-bold text-white">Deploy Application</h1>
                            <p className="text-gray-400 mt-1">Configure and deploy your application</p>
                        </div>
                    </div>
                    <div className="space-y-6">
                        {/* Environment Selection */}
                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-3">
                                Environment
                            </label>
                            <div className="grid grid-cols-3 gap-3">
                                {environments?.map((env) => (
                                    <button
                                        key={env.id}
                                        onClick={() => setSelectedEnv(env.name)}
                                        className={`p-4 rounded-lg border-2 transition ${selectedEnv === env.name
                                                ? 'border-blue-500 bg-blue-500/10'
                                                : 'border-gray-700 bg-gray-900/50 hover:border-gray-600'
                                            }`}
                                    >
                                        <div className="text-white font-medium capitalize">{env.name}</div>
                                        <div className="text-xs text-gray-400 mt-1">{env.type}</div>
                                    </button>
                                ))}
                            </div>
                        </div>
                        {/* Branch Selection */}
                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-3">
                                <GitBranch className="w-4 h-4 inline mr-2" />
                                Branch
                            </label>
                            <select
                                value={selectedBranch}
                                onChange={(e) => setSelectedBranch(e.target.value)}
                                className="w-full px-4 py-3 bg-gray-900/50 border border-gray-700 rounded-lg text-white focus:border-blue-500 focus:outline-none"
                            >
                                <option value="main">main</option>
                                <option value="develop">develop</option>
                                <option value="staging">staging</option>
                            </select>
                        </div>
                        {/* Strategy Selection */}
                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-3">
                                <Settings className="w-4 h-4 inline mr-2" />
                                Deployment Strategy
                            </label>
                            <div className="grid grid-cols-2 gap-3">
                                {['rolling', 'blue-green', 'canary', 'direct'].map((strategy) => (
                                    <button
                                        key={strategy}
                                        onClick={() => setSelectedStrategy(strategy)}
                                        className={`p-4 rounded-lg border-2 transition text-left ${selectedStrategy === strategy
                                                ? 'border-blue-500 bg-blue-500/10'
                                                : 'border-gray-700 bg-gray-900/50 hover:border-gray-600'
                                            }`}
                                    >
                                        <div className="text-white font-medium capitalize">{strategy}</div>
                                        <div className="text-xs text-gray-400 mt-1">
                                            {strategy === 'rolling' && 'Zero downtime, gradual rollout'}
                                            {strategy === 'blue-green' && 'Instant switch, easy rollback'}
                                            {strategy === 'canary' && 'Progressive traffic shift'}
                                            {strategy === 'direct' && 'Fast, simple deployment'}
                                        </div>
                                    </button>
                                ))}
                            </div>
                        </div>
                        {/* Deploy Button */}
                        <button
                            onClick={handleDeploy}
                            disabled={deployMutation.isPending}
                            className="w-full py-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white rounded-lg font-semibold text-lg transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3"
                        >
                            <Play className="w-5 h-5" />
                            {deployMutation.isPending ? 'Deploying...' : 'Deploy Now'}
                        </button>
                        {deployMutation.isSuccess && (
                            <div className="p-4 bg-green-500/10 border border-green-500/30 rounded-lg">
                                <div className="text-green-400 font-medium">✓ Deployment started successfully!</div>
                                <div className="text-sm text-gray-400 mt-1">
                                    Check the deployments page for progress
                                </div>
                            </div>
                        )}
                        {deployMutation.isError && (
                            <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
                                <div className="text-red-400 font-medium">✗ Deployment failed</div>
                                <div className="text-sm text-gray-400 mt-1">
                                    {deployMutation.error?.message || 'Unknown error occurred'}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
