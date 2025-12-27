import { createSignal, createEffect, onCleanup, Show, For } from 'solid-js';
import { OsqueryAPI, OsqueryReport } from '../../api/osquery';

const OsqueryHosts = () => {
  const [reports, setReports] = createSignal<Record<string, OsqueryReport>>({});
  const [loading, setLoading] = createSignal(true);

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
          <div class="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <For each={Object.entries(reports())}>
              {([agentId, report]) => (
                <div class="mb-4 p-4 border border-gray-200 dark:border-gray-700 rounded">
                  <h3 class="font-semibold text-lg mb-2 text-gray-800 dark:text-gray-200">{agentId}</h3>
                  <div class="text-sm text-gray-600 dark:text-gray-400">
                    <p>Processes: {report.processes?.length || 0}</p>
                    <p>Services: {report.services?.length || 0}</p>
                    <p>Last Updated: {new Date(report.timestamp).toLocaleString()}</p>
                  </div>
                </div>
              )}
            </For>
          </div>
        </div>
      </Show>
    </Show>
  );
};

export default OsqueryHosts;
