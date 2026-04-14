import SwiftUI

struct HomeView: View {
    @State private var categories: [Category] = []
    @State private var searchText = ""
    @State private var searchResults: [ContentItem] = []
    @State private var selectedURL: URL?
    @State private var searchTask: Task<Void, Never>?
    @EnvironmentObject private var syncManager: SyncManager

    private var isSearching: Bool {
        !searchText.trimmingCharacters(in: .whitespaces).isEmpty
    }

    var body: some View {
        NavigationStack {
            Group {
                if isSearching {
                    searchResultsList
                } else {
                    categoriesList
                }
            }
            .safeAreaInset(edge: .bottom) {
                SyncStatusBar(state: syncManager.state)
            }
            .navigationTitle("羽球大师课")
            .navigationDestination(for: Category.self) { category in
                CategoryView(category: category)
            }
            .searchable(text: $searchText, prompt: "搜索教程")
            .onChange(of: searchText) { _, newValue in
                searchTask?.cancel()
                searchTask = Task {
                    try? await Task.sleep(nanoseconds: 300_000_000) // 300ms debounce
                    guard !Task.isCancelled else { return }
                    let results = await Database.shared.searchContentsAsync(keyword: newValue)
                    guard !Task.isCancelled else { return }
                    await MainActor.run { searchResults = results }
                }
            }
            .sheet(item: $selectedURL) { url in
                SafariView(url: url)
                    .ignoresSafeArea()
            }
            .task {
                categories = await Database.shared.categoriesAsync(parentId: nil)
            }
        }
    }

    private var categoriesList: some View {
        Group {
            if categories.isEmpty {
                ContentUnavailableView(
                    "暂无内容",
                    systemImage: "tray",
                    description: Text("下拉刷新获取最新数据")
                )
            } else {
                List(categories) { category in
                    NavigationLink(value: category) {
                        HStack(spacing: 12) {
                            Text(category.icon)
                                .font(.title2)
                            Text(category.name)
                                .font(.body)
                        }
                        .padding(.vertical, 4)
                    }
                }
            }
        }
        .refreshable {
            await DataSync.syncDatabase()
            categories = await Database.shared.categoriesAsync(parentId: nil)
        }
    }

    private var searchResultsList: some View {
        Group {
            if searchResults.isEmpty {
                ContentUnavailableView("无搜索结果", systemImage: "magnifyingglass")
            } else {
                List(searchResults) { item in
                    Button {
                        DeepLink.open(sourceUrl: item.sourceUrl, sourcePlatform: item.sourcePlatform) { url in
                            selectedURL = url
                        }
                    } label: {
                        ContentRow(item: item)
                    }
                    .tint(.primary)
                }
            }
        }
    }
}

// MARK: - Sync Status Bar

private struct SyncStatusBar: View {
    let state: SyncState

    var body: some View {
        Group {
            switch state {
            case .idle:
                EmptyView()
            case .syncing:
                HStack(spacing: 6) {
                    ProgressView()
                        .controlSize(.small)
                    Text("正在同步...")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
                .frame(maxWidth: .infinity)
                .padding(.vertical, 6)
                .background(.ultraThinMaterial)
            case .success:
                Text("已同步")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 6)
                    .background(.ultraThinMaterial)
                    .transition(.opacity)
            case .failed:
                Text("同步失败")
                    .font(.caption)
                    .foregroundStyle(.red)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 6)
                    .background(.ultraThinMaterial)
                    .transition(.opacity)
            }
        }
        .animation(.easeInOut(duration: 0.3), value: state)
    }
}
