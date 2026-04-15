import SwiftUI

struct HomeView: View {
    @State private var categories: [Category] = []
    @State private var learningPaths: [LearningPath] = []
    @State private var searchText = ""
    @State private var searchResults: [ContentItem] = []
    @State private var selectedURL: URL?
    @State private var searchTask: Task<Void, Never>?
    @State private var favoriteItems: [ContentItem] = []
    @EnvironmentObject private var syncManager: SyncManager
    @EnvironmentObject private var userState: UserState

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
            .navigationDestination(for: LearningPath.self) { path in
                PathDetailView(path: path)
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
                async let cats = Database.shared.categoriesAsync(parentId: nil)
                async let paths = Database.shared.learningPathsAsync()
                categories = await cats
                learningPaths = await paths
                favoriteItems = await Database.shared.contentsByIdsAsync(userState.favorites)
            }
            .onChange(of: userState.favorites) { _, newFavorites in
                Task {
                    favoriteItems = await Database.shared.contentsByIdsAsync(newFavorites)
                }
            }
        }
    }

    private var categoriesList: some View {
        Group {
            if categories.isEmpty && learningPaths.isEmpty {
                ContentUnavailableView(
                    "暂无内容",
                    systemImage: "tray",
                    description: Text("下拉刷新获取最新数据")
                )
            } else {
                List {
                    if !learningPaths.isEmpty {
                        Section {
                            ScrollView(.horizontal, showsIndicators: false) {
                                HStack(spacing: 12) {
                                    ForEach(learningPaths) { path in
                                        NavigationLink(value: path) {
                                            PathCard(path: path)
                                        }
                                        .buttonStyle(.plain)
                                    }
                                }
                                .padding(.horizontal, 16)
                                .padding(.vertical, 4)
                            }
                            .listRowInsets(EdgeInsets())
                            .listRowSeparator(.hidden)
                        } header: {
                            Text("学习路径")
                                .font(.title2)
                                .fontWeight(.bold)
                        }
                    }

                    if !favoriteItems.isEmpty {
                        Section {
                            ForEach(favoriteItems) { item in
                                Button {
                                    DeepLink.open(sourceUrl: item.sourceUrl, sourcePlatform: item.sourcePlatform) { url in
                                        selectedURL = url
                                    }
                                } label: {
                                    ContentRow(item: item)
                                }
                                .tint(.primary)
                            }
                        } header: {
                            Text("我的收藏")
                                .font(.title2)
                                .fontWeight(.bold)
                        }
                    }

                    Section {
                        ForEach(categories) { category in
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
            }
        }
        .refreshable {
            await DataSync.syncDatabase()
            async let cats = Database.shared.categoriesAsync(parentId: nil)
            async let paths = Database.shared.learningPathsAsync()
            categories = await cats
            learningPaths = await paths
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

// MARK: - Path Card

private struct PathCard: View {
    let path: LearningPath

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(path.title)
                    .font(.headline)
                    .foregroundStyle(Color(red: 0x11/255.0, green: 0x11/255.0, blue: 0x11/255.0))
                    .lineLimit(1)

                Spacer()

                DifficultyBadge(difficulty: path.difficulty)
            }

            if !path.summary.isEmpty {
                Text(path.summary)
                    .font(.subheadline)
                    .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                    .lineLimit(2)
            }

            Text("\(path.stepCount) 步")
                .font(.caption)
                .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
        }
        .padding(16)
        .frame(width: 220, alignment: .leading)
        .background(Color(red: 0xF5/255.0, green: 0xF5/255.0, blue: 0xF5/255.0))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

private struct DifficultyBadge: View {
    let difficulty: String

    private var displayName: String {
        switch difficulty {
        case "beginner": return "入门"
        case "intermediate": return "进阶"
        case "advanced": return "高级"
        default: return difficulty
        }
    }

    private var badgeColor: Color {
        switch difficulty {
        case "beginner": return Color(red: 0x00/255.0, green: 0x7D/255.0, blue: 0x48/255.0)       // Success Green
        case "intermediate": return Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0)    // Secondary
        case "advanced": return Color(red: 0xD3/255.0, green: 0x00/255.0, blue: 0x05/255.0)        // Error Red
        default: return .gray
        }
    }

    var body: some View {
        Text(displayName)
            .font(.caption2)
            .fontWeight(.medium)
            .padding(.horizontal, 8)
            .padding(.vertical, 2)
            .background(badgeColor.opacity(0.15))
            .foregroundStyle(badgeColor)
            .clipShape(Capsule())
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
