import { createSignal, createEffect, onCleanup, Show, For } from 'solid-js';
import { OsqueryAPI, OsqueryReport } from '../../api/osquery';

const OsqueryHosts = () => {
  const [reports, setReports] = createSignal<Record<string, OsqueryReport>>({});
  const [loading, setLoading] = createSignal(true);
  const [searchTerm, setSearchTerm] = createSignal('');
  const [selectedAgent, setSelectedAgent] = createSignal<string | null>(null);

  const loadReports = async () => {
    try {
      const data = await OsqueryAPI.getAllReports();
      setReports(data);
    } catch (error) {
      console.error('Failed to load osquery reports:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatMemory = (bytes?: string) => {
    if (!bytes) return 'N/A';
    const num = parseInt(bytes);
    if (isNaN(num)) return 'N/A';
    if (num < 1024) return `${num} B`;
    if (num < 1024 * 1024) return `${(num / 1024).toFixed(1)} KB`;
    if (num < 1024 * 1024 * 1024) return `${(num / (1024 * 1024)).toFixed(1)} MB`;
    return `${(num / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  };

  const calculateTotalMemory = (processes: OsqueryReport['processes']) => {
    const total = processes.reduce((sum, p) => {
      const mem = parseInt(p.memory_bytes || '0');
      return sum + (isNaN(mem) ? 0 : mem);
    }, 0);
    return formatMemory(total.toString());
  };

  const allProcesses = () => {
    const agent = selectedAgent();
    if (!agent) return [];
    const report = reports()[agent];
    if (!report) return [];
    return report.processes
      .filter((p) => p.name.toLowerCase().includes(searchTerm().toLowerCase()))
      .map((p) => ({ ...p, agentId: agent }));
  };

  createEffect(() => {
    loadReports();
    const interval = setInterval(loadReports, 30000);
    onCleanup(() => clearInterval(interval));
  });

  createEffect(() => {
    const agentIds = Object.keys(reports());
    if (agentIds.length > 0 && !selectedAgent()) {
      setSelectedAgent(agentIds[0]);
    }
  });

  return (
    <Show
      when={!loading()}
      fallback={<div class="p-8 text-center">Loading osquery data...</div>}
    >
      <Show
        when={Object.keys(reports()).length > 0}
        fallback={<div class="p-8 text-center text-gray-500 dark:text-gray-400">No osquery agents reporting</div>}
      >
        <div class="p-6">
          <h1 class="text-2xl font-bold mb-6 text-gray-800 dark:text-gray-200">osquery Monitoring</h1>
          
          {/* Agents */}
          <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm mb-6 overflow-hidden border border-gray-200 dark:border-gray-700">
            <h2 class="text-xl font-semibold p-4 bg-gray-50 dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">Agents</h2>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead class="bg-gray-50 dark:bg-gray-900">
                  <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Agent ID</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Processes</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Services</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Total Memory</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Last Update</th>
                  </tr>
                </thead>
                <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  <For each={Object.entries(reports())}>
                    {([agentId, report]) => (
                      <tr 
                        class="hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer transition-colors"
                        classList={{ 'bg-blue-50 dark:bg-blue-900/30': selectedAgent() === agentId }}
                        onClick={() => setSelectedAgent(agentId)}
                      >
                        <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100">{agentId}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{report.processes?.length || 0}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{report.services?.length || 0}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{calculateTotalMemory(report.processes)}</td>
                        <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{new Date(report.timestamp).toLocaleString()}</td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </div>

          {/* Running Processes */}
          <Show when={selectedAgent()}>
            <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm overflow-hidden border border-gray-200 dark:border-gray-700">
              <div class="p-4 bg-gray-50 dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
                <h2 class="text-xl font-semibold mb-3 text-gray-900 dark:text-gray-100">Processes - {selectedAgent()}</h2>
                <input
                  type="text"
                  placeholder="Search processes..."
                  class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  value={searchTerm()}
                  onInput={(e) => setSearchTerm(e.currentTarget.value)}
                />
              </div>
              <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                  <thead class="bg-gray-50 dark:bg-gray-900">
                    <tr>
                      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">PID</th>
                      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Name</th>
                      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Path</th>
                      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Status</th>
                      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Memory</th>
                      <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Log Files</th>
                    </tr>
                  </thead>
                  <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                    <For each={allProcesses()}>
                      {(process) => (
                        <tr class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">{process.pid}</td>
                          <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-gray-100">{process.name}</td>
                          <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400 max-w-md truncate">{process.path}</td>
                          <td class="px-6 py-4 whitespace-nowrap text-sm">
                            <span class="px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                              {process.status || 'running'}
                            </span>
                          </td>
                          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">{formatMemory(process.memory_bytes)}</td>
                          <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                            <Show when={process.log_files?.length > 0} fallback={
                              <Show when={process.log_command} fallback={<span class="text-gray-400">None</span>}>
                                <code class="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded font-mono">{process.log_command}</code>
                              </Show>
                            }>
                              <ul class="list-disc list-inside space-y-1">
                                <For each={process.log_files.slice(0, 3)}>
                                  {(log) => <li class="truncate max-w-xs" title={log}>{log}</li>}
                                </For>
                                <Show when={process.log_files.length > 3}>
                                  <li class="text-gray-400">+{process.log_files.length - 3} more</li>
                                </Show>
                              </ul>
                            </Show>
                          </td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </div>
            </div>
          </Show>
        </div>
      </Show>
    </Show>
  );
};

export default OsqueryHosts;
