import { createSignal, createEffect, onCleanup, Show, For } from 'solid-js';
import { OsqueryAPI, OsqueryReport } from '../../api/osquery';

const OsqueryHosts = () => {
  const [reports, setReports] = createSignal<Record<string, OsqueryReport>>({});
  const [loading, setLoading] = createSignal(true);
  const [searchTerm, setSearchTerm] = createSignal('');

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
    return Object.entries(reports()).flatMap(([agentId, report]) =>
      report.processes.map((p) => ({ ...p, agentId }))
    ).filter((p) => p.name.toLowerCase().includes(searchTerm().toLowerCase()));
  };

  createEffect(() => {
    loadReports();
    const interval = setInterval(loadReports, 30000);
    onCleanup(() => clearInterval(interval));
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
          
          {/* Host Summary */}
          <div class="bg-white dark:bg-gray-800 rounded-lg shadow mb-6 overflow-hidden">
            <h2 class="text-xl font-semibold p-4 border-b border-gray-200 dark:border-gray-700">Host Summary</h2>
            <div class="overflow-x-auto">
              <table class="w-full">
                <thead class="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th class="px-4 py-2 text-left text-sm font-medium">Agent ID</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Processes</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Services</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Total Memory</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Last Update</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                  <For each={Object.entries(reports())}>
                    {([agentId, report]) => (
                      <tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
                        <td class="px-4 py-2">{agentId}</td>
                        <td class="px-4 py-2">{report.processes?.length || 0}</td>
                        <td class="px-4 py-2">{report.services?.length || 0}</td>
                        <td class="px-4 py-2">{calculateTotalMemory(report.processes)}</td>
                        <td class="px-4 py-2">{new Date(report.timestamp).toLocaleString()}</td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </div>

          {/* Running Processes */}
          <div class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <div class="p-4 border-b border-gray-200 dark:border-gray-700">
              <h2 class="text-xl font-semibold mb-2">Running Processes</h2>
              <input
                type="text"
                placeholder="Search processes..."
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700"
                value={searchTerm()}
                onInput={(e) => setSearchTerm(e.currentTarget.value)}
              />
            </div>
            <div class="overflow-x-auto">
              <table class="w-full">
                <thead class="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th class="px-4 py-2 text-left text-sm font-medium">PID</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Name</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Path</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Memory</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Log Files</th>
                    <th class="px-4 py-2 text-left text-sm font-medium">Agent</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                  <For each={allProcesses()}>
                    {(process) => (
                      <tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
                        <td class="px-4 py-2 text-sm">{process.pid}</td>
                        <td class="px-4 py-2 font-medium">{process.name}</td>
                        <td class="px-4 py-2 text-sm text-gray-600 dark:text-gray-400">{process.path}</td>
                        <td class="px-4 py-2 text-sm">{formatMemory(process.memory_bytes)}</td>
                        <td class="px-4 py-2 text-sm">
                          <Show when={process.log_files?.length > 0} fallback={<span class="text-gray-400">None</span>}>
                            <ul class="list-disc list-inside">
                              <For each={process.log_files.slice(0, 3)}>
                                {(log) => <li class="truncate max-w-xs" title={log}>{log}</li>}
                              </For>
                              <Show when={process.log_files.length > 3}>
                                <li>+{process.log_files.length - 3} more</li>
                              </Show>
                            </ul>
                          </Show>
                        </td>
                        <td class="px-4 py-2">{process.agentId}</td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </Show>
    </Show>
  );
};

export default OsqueryHosts;
