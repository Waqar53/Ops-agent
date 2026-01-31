'use client';
import { useParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { DollarSign, TrendingDown, TrendingUp, Zap, Server, Database, HardDrive } from 'lucide-react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
export default function CostPage() {
    const params = useParams();
    const projectId = params.id as string;
    const { data: costData } = useQuery({
        queryKey: ['cost', projectId],
        queryFn: () => api.getCost(projectId),
    });
    const { data: forecast } = useQuery({
        queryKey: ['cost-forecast', projectId],
        queryFn: () => api.getCostForecast(projectId),
    });
    const breakdown = [
        { name: 'Compute', value: 450, color: '#3B82F6' },
        { name: 'Database', value: 380, color: '#8B5CF6' },
        { name: 'Storage', value: 120, color: '#10B981' },
        { name: 'Network', value: 180, color: '#F59E0B' },
        { name: 'Other', value: 117, color: '#6B7280' },
    ];
    const recommendations = [
        {
            title: 'Use Spot Instances',
            description: 'Save 70% on compute costs for non-critical workloads',
            savings: 315,
            icon: Server,
        },
        {
            title: 'Right-size Databases',
            description: 'Current database is over-provisioned by 40%',
            savings: 152,
            icon: Database,
        },
        {
            title: 'Enable Storage Tiering',
            description: 'Move infrequently accessed data to cheaper storage',
            savings: 48,
            icon: HardDrive,
        },
    ];
    return (
        <div className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
            <div className="max-w-7xl mx-auto px-6 py-8">
                {/* Current Cost */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-gray-400 text-sm">Current Month</span>
                            <DollarSign className="w-5 h-5 text-green-500" />
                        </div>
                        <div className="text-4xl font-bold text-white mb-1">$1,247</div>
                        <div className="flex items-center gap-2 text-sm text-green-500">
                            <TrendingDown className="w-4 h-4" />
                            -$183 from last month
                        </div>
                    </div>
                    <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-gray-400 text-sm">Forecast (End of Month)</span>
                            <TrendingUp className="w-5 h-5 text-blue-500" />
                        </div>
                        <div className="text-4xl font-bold text-white mb-1">$1,580</div>
                        <div className="text-sm text-gray-400">
                            Based on current usage trends
                        </div>
                    </div>
                    <div className="bg-gradient-to-br from-green-500/20 to-emerald-500/20 border border-green-500/30 rounded-xl p-6">
                        <div className="flex items-center justify-between mb-2">
                            <span className="text-green-400 text-sm font-medium">Potential Savings</span>
                            <Zap className="w-5 h-5 text-green-400" />
                        </div>
                        <div className="text-4xl font-bold text-white mb-1">$515</div>
                        <div className="text-sm text-green-400">
                            32% cost reduction available
                        </div>
                    </div>
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                    {/* Cost Breakdown */}
                    <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                        <h2 className="text-xl font-bold text-white mb-6">Cost Breakdown</h2>
                        <ResponsiveContainer width="100%" height={300}>
                            <PieChart>
                                <Pie
                                    data={breakdown}
                                    cx="50%"
                                    cy="50%"
                                    labelLine={false}
                                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                                    outerRadius={100}
                                    fill="#8884d8"
                                    dataKey="value"
                                >
                                    {breakdown.map((entry, index) => (
                                        <Cell key={`cell-${index}`} fill={entry.color} />
                                    ))}
                                </Pie>
                                <Tooltip
                                    contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
                                    labelStyle={{ color: '#F3F4F6' }}
                                />
                            </PieChart>
                        </ResponsiveContainer>
                        <div className="grid grid-cols-2 gap-3 mt-6">
                            {breakdown.map((item) => (
                                <div key={item.name} className="flex items-center gap-2">
                                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }}></div>
                                    <span className="text-sm text-gray-400">{item.name}</span>
                                    <span className="text-sm text-white font-medium ml-auto">${item.value}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                    {/* Optimization Recommendations */}
                    <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                        <h2 className="text-xl font-bold text-white mb-6">Optimization Recommendations</h2>
                        <div className="space-y-4">
                            {recommendations.map((rec, index) => (
                                <div key={index} className="bg-gray-900/50 border border-gray-700 rounded-lg p-4 hover:border-green-500/50 transition group">
                                    <div className="flex items-start gap-4">
                                        <div className="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center shrink-0">
                                            <rec.icon className="w-5 h-5 text-green-400" />
                                        </div>
                                        <div className="flex-1">
                                            <h3 className="text-white font-medium mb-1">{rec.title}</h3>
                                            <p className="text-sm text-gray-400 mb-2">{rec.description}</p>
                                            <div className="flex items-center justify-between">
                                                <span className="text-sm text-green-400 font-medium">
                                                    Save ${rec.savings}/month
                                                </span>
                                                <button className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white text-sm rounded transition opacity-0 group-hover:opacity-100">
                                                    Apply
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
                {/* Detailed Breakdown */}
                <div className="bg-gray-800/50 backdrop-blur-sm border border-gray-700 rounded-xl p-6">
                    <h2 className="text-xl font-bold text-white mb-6">Detailed Cost Breakdown</h2>
                    <div className="overflow-x-auto">
                        <table className="w-full">
                            <thead>
                                <tr className="border-b border-gray-700">
                                    <th className="text-left py-3 px-4 text-gray-400 font-medium">Resource</th>
                                    <th className="text-left py-3 px-4 text-gray-400 font-medium">Type</th>
                                    <th className="text-right py-3 px-4 text-gray-400 font-medium">Usage</th>
                                    <th className="text-right py-3 px-4 text-gray-400 font-medium">Cost</th>
                                </tr>
                            </thead>
                            <tbody>
                                {[
                                    { name: 'web-server-prod', type: 'EC2 t3.medium', usage: '720 hours', cost: '$30.24' },
                                    { name: 'api-server-prod', type: 'EC2 t3.large', usage: '720 hours', cost: '$60.48' },
                                    { name: 'postgres-prod', type: 'RDS db.t3.medium', usage: '720 hours', cost: '$51.84' },
                                    { name: 'redis-cache', type: 'ElastiCache t3.micro', usage: '720 hours', cost: '$12.96' },
                                    { name: 'load-balancer', type: 'ALB', usage: '720 hours', cost: '$16.20' },
                                    { name: 's3-storage', type: 'S3 Standard', usage: '500 GB', cost: '$11.50' },
                                ].map((resource, index) => (
                                    <tr key={index} className="border-b border-gray-800 hover:bg-gray-900/30">
                                        <td className="py-3 px-4 text-white">{resource.name}</td>
                                        <td className="py-3 px-4 text-gray-400">{resource.type}</td>
                                        <td className="py-3 px-4 text-gray-400 text-right">{resource.usage}</td>
                                        <td className="py-3 px-4 text-white font-medium text-right">{resource.cost}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
}
