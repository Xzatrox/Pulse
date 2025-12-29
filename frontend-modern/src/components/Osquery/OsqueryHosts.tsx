import { createSignal, createEffect, onCleanup, Show, For, createMemo } from 'solid-js';
import { OsqueryAPI, OsqueryReport } from '../../api/osquery';
import { Card } from '@/components/shared/Card';
import { ScrollableTable } from '@/components/shared/ScrollableTable';
import { useBreakpoint } from '@/hooks/useBreakpoint';

type SortKey = 'name' | 'pid' | 'memory' | 'status';
type SortDirection = 'asc' | 'desc';

const OsqueryHosts = () => {
  const [reports, setReports] = createSignal<Record<string, OsqueryReport>>({});
  const [loading, setLoading] = createSignal(true);
  const [searchTerm, setSearchTerm] = createSignal('');
  const [selectedAgent, setSelectedAgent] = createSignal<string | null>(null);
  const [sortKey, setSortKey] = createSignal<SortKey>('name');
  const [sortDirection, setSortDirection] = createSignal<SortDirection>('asc');
  const { isMobile } = useBreakpoint();

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

  const handleSort = (key: SortKey) => {
    if (sortKey() === key) {
      setSortDirection(sortDirection() === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDirection('asc');
    }
  };

  const renderSortIndicator = (key: SortKey) => {
    if (sortKey() !== key) return null;
    return sortDirection() === 'asc' ? '▲' : '▼';
  };

  const sortedProcesses = createMemo(() => {
    const agent = selectedAgent();
    if (!agent) return [];
    const report = reports()[agent];
    if (!report) return [];
    
    const filtered = report.processes
      .filter((p) => p.name.toLowerCase().includes(searchTerm().toLowerCase()))
      .map((p) => ({ ...p, agentId: agent }));

    const key = sortKey();
    const dir = sortDirection();

    filtered.sort((a, b) => {
      let value = 0;
      switch (key) {
        case 'name':
          value = a.name.localeCompare(b.name);
          break;
        case 'pid':
          value = parseInt(a.pid) - parseInt(b.pid);
          break;
        case 'memory':
          value = parseInt(a.memory_bytes || '0') - parseInt(b.memory_bytes || '0');
          break;
        case 'status':
          value = (a.status || 'running').localeCompare(b.status || 'running');
          break;
      }
      return dir === 'asc' ? value : -value;
    });

    return filtered;
  });

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
          <h1 class="text-2xl font-bold mb-6 text-gray-800 dark:text-gray-200">Services Monitoring</h1>
          
          {/* Agents */}
          <Card padding="none" tone="glass" class="mb-4 overflow-hidden">
            <ScrollableTable persistKey="osquery-agents" minWidth={isMobile() ? '100%' : '600px'}>
              <table class="w-full border-collapse whitespace-nowrap">
                <thead>
                  <tr class="bg-gray-50 dark:bg-gray-700/50 text-gray-600 dark:text-gray-300 border-b border-gray-200 dark:border-gray-700">
                    <th class="pl-3 pr-2 py-1 text-left text-[11px] sm:text-xs font-medium uppercase tracking-wider whitespace-nowrap">Agent ID</th>
                    <th class="px-2 py-1 text-center text-[11px] sm:text-xs font-medium uppercase tracking-wider whitespace-nowrap">Processes</th>
                    <th class="px-2 py-1 text-center text-[11px] sm:text-xs font-medium uppercase tracking-wider whitespace-nowrap">Services</th>
                    <th class="px-2 py-1 text-center text-[11px] sm:text-xs font-medium uppercase tracking-wider whitespace-nowrap">Total Memory</th>
                    <th class="pr-3 pl-2 py-1 text-right text-[11px] sm:text-xs font-medium uppercase tracking-wider whitespace-nowrap">Last Update</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                  <For each={Object.entries(reports())}>
                    {([agentId, report]) => (
                      <tr 
                        class="hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer transition-colors"
                        classList={{ 'bg-blue-50 dark:bg-blue-900/30': selectedAgent() === agentId }}
                        onClick={() => setSelectedAgent(agentId)}
                      >
                        <td class="pl-3 pr-2 py-2 text-sm font-medium text-gray-900 dark:text-gray-100">{agentId}</td>
                        <td class="px-2 py-2 text-sm text-center text-gray-500 dark:text-gray-400">{report.processes?.length || 0}</td>
                        <td class="px-2 py-2 text-sm text-center text-gray-500 dark:text-gray-400">{report.services?.length || 0}</td>
                        <td class="px-2 py-2 text-sm text-center text-gray-500 dark:text-gray-400">{calculateTotalMemory(report.processes)}</td>
                        <td class="pr-3 pl-2 py-2 text-sm text-right text-gray-500 dark:text-gray-400">{new Date(report.timestamp).toLocaleString()}</td>
                      </tr>
                    )}
                  </For>
                </tbody>
              </table>
            </ScrollableTable>
          </Card>

          {/* Running Processes */}
          <Show when={selectedAgent()}>
            <Card padding="none" tone="glass" class="overflow-hidden">
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
              <ScrollableTable persistKey="osquery-processes" minWidth={isMobile() ? '100%' : '900px'}>
                <table class="w-full border-collapse whitespace-nowrap">
                  <thead>
                    <tr class="bg-gray-50 dark:bg-gray-700/50 text-gray-600 dark:text-gray-300 border-b border-gray-200 dark:border-gray-700">
                      <th 
                        class="pl-3 pr-2 py-1 text-left text-[11px] sm:text-xs font-medium uppercase tracking-wider cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600 whitespace-nowrap"
                        onClick={() => handleSort('pid')}
                      >
                        PID {renderSortIndicator('pid')}
                      </th>
                      <th 
                        class="px-2 py-1 text-left text-[11px] sm:text-xs font-medium uppercase tracking-wider cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600 whitespace-nowrap"
                        onClick={() => handleSort('name')}
                      >
                        Name {renderSortIndicator('name')}
                      </th>
                      <th class="px-2 py-1 text-left text-[11px] sm:text-xs font-medium uppercase tracking-wider whitespace-nowrap">Path</th>
                      <th 
                        class="px-2 py-1 text-center text-[11px] sm:text-xs font-medium uppercase tracking-wider cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600 whitespace-nowrap"
                        onClick={() => handleSort('status')}
                      >
                        Status {renderSortIndicator('status')}
                      </th>
                      <th 
                        class="px-2 py-1 text-right text-[11px] sm:text-xs font-medium uppercase tracking-wider cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600 whitespace-nowrap"
                        onClick={() => handleSort('memory')}
                      >
                        Memory {renderSortIndicator('memory')}
                      </th>
                      <th class="pr-3 pl-2 py-1 text-left text-[11px] sm:text-xs font-medium uppercase tracking-wider whitespace-nowrap">Log Files</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                    <For each={sortedProcesses()}>
                      {(process) => (
                        <tr class="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
                          <td class="pl-3 pr-2 py-2 text-sm text-gray-900 dark:text-gray-100">{process.pid}</td>
                          <td class="px-2 py-2 text-sm font-medium text-gray-900 dark:text-gray-100">{process.name}</td>
                          <td class="px-2 py-2 text-sm text-gray-500 dark:text-gray-400 max-w-md truncate">{process.path}</td>
                          <td class="px-2 py-2 text-sm text-center">
                            <span class="inline-flex px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                              {process.status || 'running'}
                            </span>
                          </td>
                          <td class="px-2 py-2 text-sm text-right text-gray-500 dark:text-gray-400">{formatMemory(process.memory_bytes)}</td>
                          <td class="pr-3 pl-2 py-2 text-sm text-gray-500 dark:text-gray-400">
                            <Show when={process.log_files?.length > 0} fallback={
                              <Show when={process.log_command} fallback={<span class="text-gray-400">None</span>}>
                                <code class="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded font-mono">{process.log_command}</code>
                              </Show>
                            }>
                              <div class="max-w-xs">
                                <For each={process.log_files.slice(0, 2)}>
                                  {(log) => <div class="truncate" title={log}>{log}</div>}
                                </For>
                                <Show when={process.log_files.length > 2}>
                                  <div class="text-gray-400">+{process.log_files.length - 2} more</div>
                                </Show>
                              </div>
                            </Show>
                          </td>
                        </tr>
                      )}
                    </For>
                  </tbody>
                </table>
              </ScrollableTable>
            </Card>
          </Show>
        </div>
      </Show>
    </Show>
  );
};

export default OsqueryHosts;
