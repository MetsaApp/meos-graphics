package web

import (
	"html/template"
	"time"
)

// GetTemplates returns all HTML templates
func GetTemplates() *template.Template {
	tmpl := template.New("")

	// Add template functions
	tmpl.Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("15:04:05")
		},
		"formatDuration": func(d *string) string {
			if d == nil {
				return "-"
			}
			return *d
		},
		"hasValue": func(s *string) bool {
			return s != nil && *s != ""
		},
	})

	// Base layout
	tmpl = template.Must(tmpl.Parse(`
{{define "layout"}}
<!DOCTYPE html>
<html lang="en" class="h-full">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - MeOS Graphics</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="h-full bg-gray-50">
    <div class="min-h-full">
        <nav class="bg-white shadow">
            <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
                <div class="flex h-16 justify-between">
                    <div class="flex">
                        <div class="flex flex-shrink-0 items-center">
                            <h1 class="text-xl font-bold">MeOS Graphics</h1>
                        </div>
                    </div>
                    <div class="flex items-center space-x-4">
                        <a href="/docs" class="text-sm text-blue-600 hover:text-blue-800">API Documentation</a>
                        <span id="connection-status" class="text-sm text-gray-500">
                            <span class="inline-block h-2 w-2 rounded-full bg-gray-400"></span>
                            Disconnected
                        </span>
                    </div>
                </div>
            </div>
        </nav>
        <main>
            {{template "content" .}}
        </main>
    </div>
    
    <script>
        // Initialize SSE connection
        document.addEventListener('DOMContentLoaded', function() {
            const evtSource = new EventSource('/sse');
            
            evtSource.onopen = function() {
                console.log('SSE connection opened');
                document.getElementById('connection-status').innerHTML = 
                    '<span class="inline-block h-2 w-2 rounded-full bg-green-400"></span> Connected';
            };
            
            evtSource.onerror = function(err) {
                console.error('SSE error:', err);
                document.getElementById('connection-status').innerHTML = 
                    '<span class="inline-block h-2 w-2 rounded-full bg-red-400"></span> Disconnected';
            };
            
            evtSource.addEventListener('connected', function(e) {
                console.log('Connected event:', e.data);
                document.getElementById('connection-status').innerHTML = 
                    '<span class="inline-block h-2 w-2 rounded-full bg-green-400"></span> Connected';
            });
            
            evtSource.addEventListener('update', function(e) {
                console.log('Update event:', e.data);
                // Trigger HTMX refresh on all elements with refresh-data trigger
                htmx.trigger(document.body, 'refresh-data');
            });
            
            evtSource.addEventListener('heartbeat', function(e) {
                console.log('Heartbeat:', e.data);
            });
        });
    </script>
</body>
</html>
{{end}}
`))

	// Index page
	tmpl = template.Must(tmpl.Parse(`
{{define "index"}}
{{template "layout" .}}
{{end}}

{{define "content"}}
{{if eq .Title "MeOS Graphics"}}
<div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">
        <h2 class="text-2xl font-bold mb-6">Competition Classes</h2>
        
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {{range .Classes}}
            <a href="/web/classes/{{.ID}}" 
               class="block p-6 bg-white rounded-lg shadow hover:shadow-md transition-shadow">
                <h3 class="text-lg font-semibold mb-2">{{.Name}}</h3>
                <p class="text-sm text-gray-600">Class ID: {{.ID}}</p>
            </a>
            {{end}}
        </div>
        
        {{if not .Classes}}
        <div class="text-center py-12">
            <p class="text-gray-500">No classes available</p>
        </div>
        {{end}}
    </div>
</div>
{{else if .ClassName}}
<div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">
        <div class="mb-6">
            <a href="/web" class="text-blue-600 hover:text-blue-800">← Back to Classes</a>
        </div>
        
        <h2 class="text-2xl font-bold mb-6">{{.ClassName}}</h2>
        
        <div class="border-b border-gray-200">
            <nav class="-mb-px flex space-x-8" aria-label="Tabs">
                <button onclick="showTab('startlist')" id="tab-startlist"
                        class="tab-button border-b-2 border-blue-500 py-2 px-1 text-sm font-medium text-blue-600">
                    Start List
                </button>
                <button onclick="showTab('results')" id="tab-results"
                        class="tab-button border-b-2 border-transparent py-2 px-1 text-sm font-medium text-gray-500 hover:text-gray-700 hover:border-gray-300">
                    Results
                </button>
                <button onclick="showTab('splits')" id="tab-splits"
                        class="tab-button border-b-2 border-transparent py-2 px-1 text-sm font-medium text-gray-500 hover:text-gray-700 hover:border-gray-300">
                    Splits
                </button>
            </nav>
        </div>
        
        <div class="mt-6">
            <div id="content-startlist" class="tab-content"
                 hx-get="/web/classes/{{.ClassID}}/startlist"
                 hx-trigger="load, refresh-data from:body"
                 hx-target="this">
                <div class="animate-pulse">
                    <div class="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
                    <div class="h-4 bg-gray-200 rounded w-1/2 mb-4"></div>
                    <div class="h-4 bg-gray-200 rounded w-1/3"></div>
                </div>
            </div>
            
            <div id="content-results" class="tab-content hidden"
                 hx-get="/web/classes/{{.ClassID}}/results"
                 hx-trigger="revealed, refresh-data from:body"
                 hx-target="this">
            </div>
            
            <div id="content-splits" class="tab-content hidden"
                 hx-get="/web/classes/{{.ClassID}}/splits"
                 hx-trigger="revealed, refresh-data from:body"
                 hx-target="this">
            </div>
        </div>
    </div>
</div>

<script>
function showTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(el => {
        el.classList.add('hidden');
    });
    
    // Reset all tab buttons
    document.querySelectorAll('.tab-button').forEach(el => {
        el.classList.remove('border-blue-500', 'text-blue-600');
        el.classList.add('border-transparent', 'text-gray-500');
    });
    
    // Show selected tab content
    document.getElementById('content-' + tabName).classList.remove('hidden');
    
    // Highlight selected tab button
    const tabButton = document.getElementById('tab-' + tabName);
    tabButton.classList.remove('border-transparent', 'text-gray-500');
    tabButton.classList.add('border-blue-500', 'text-blue-600');
    
    // Trigger HTMX to load content if needed
    htmx.trigger('#content-' + tabName, 'revealed');
}
</script>
{{else if .Error}}
<div class="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
    <div class="px-4 py-6 sm:px-0">
        <div class="text-center">
            <h2 class="text-2xl font-bold text-red-600 mb-4">Error</h2>
            <p class="text-gray-600">{{.Error}}</p>
            <a href="/web" class="mt-4 inline-block text-blue-600 hover:text-blue-800">← Back to Home</a>
        </div>
    </div>
</div>
{{end}}
{{end}}
`))

	// Class page
	tmpl = template.Must(tmpl.Parse(`
{{define "class"}}
{{template "layout" .}}
{{end}}
`))

	// Start list partial
	tmpl = template.Must(tmpl.Parse(`
{{define "startlist-partial"}}
<div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
    <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
            <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Club</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Start Time</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Card</th>
            </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
            {{range .StartList}}
            <tr>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{.Name}}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{.Club}}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{formatTime .StartTime}}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{if .Card}}{{.Card}}{{else}}-{{end}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{if not .StartList}}
    <div class="text-center py-12">
        <p class="text-gray-500">No competitors in start list</p>
    </div>
    {{end}}
</div>
{{end}}
`))

	// Results partial
	tmpl = template.Must(tmpl.Parse(`
{{define "results-partial"}}
<div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
    <table class="min-w-full divide-y divide-gray-300">
        <thead class="bg-gray-50">
            <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Pos</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Club</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Behind</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
            {{range .Results}}
            <tr class="{{if eq .Status "DNF"}}bg-red-50{{else if eq .Status "DNS"}}bg-gray-50{{else if eq .Status "MP"}}bg-yellow-50{{else if eq .Status "Running"}}bg-blue-50{{end}}">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{if .Position}}{{.Position}}{{else}}-{{end}}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{.Name}}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{.Club}}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {{if hasValue .Time}}{{formatDuration .Time}}{{else}}-{{end}}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {{formatDuration .TimeDifference}}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm">
                    <span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5
                        {{if eq .Status "OK"}}bg-green-100 text-green-800{{else if eq .Status "DNF"}}bg-red-100 text-red-800{{else if eq .Status "DNS"}}bg-gray-100 text-gray-800{{else if eq .Status "MP"}}bg-yellow-100 text-yellow-800{{else if eq .Status "Running"}}bg-blue-100 text-blue-800{{end}}">
                        {{.Status}}
                    </span>
                </td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{if not .Results}}
    <div class="text-center py-12">
        <p class="text-gray-500">No results available</p>
    </div>
    {{end}}
</div>
{{end}}
`))

	// Splits partial
	tmpl = template.Must(tmpl.Parse(`
{{define "splits-partial"}}
<div class="space-y-6">
    {{range .Splits.Splits}}
    <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
        <div class="bg-gray-50 px-6 py-3">
            <h3 class="text-sm font-medium text-gray-900">{{.ControlName}}</h3>
        </div>
        <table class="min-w-full divide-y divide-gray-300">
            <thead class="bg-gray-50">
                <tr>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Pos</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Club</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
                    <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Behind</th>
                </tr>
            </thead>
            <tbody class="bg-white divide-y divide-gray-200">
                {{range .Standings}}
                <tr>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {{if .Position}}{{.Position}}{{else}}-{{end}}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{{.Name}}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{{.Club}}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {{formatDuration .ElapsedTime}}
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {{formatDuration .TimeDifference}}
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    {{end}}
</div>
{{end}}
`))

	// Error pages
	tmpl = template.Must(tmpl.Parse(`
{{define "error"}}
{{template "layout" .}}
{{end}}

{{define "error-partial"}}
<div class="bg-red-50 border border-red-200 rounded-md p-4">
    <p class="text-red-800">{{.Error}}</p>
</div>
{{end}}
`))

	return tmpl
}
