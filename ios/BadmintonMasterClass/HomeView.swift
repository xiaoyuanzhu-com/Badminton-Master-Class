import SwiftUI

struct HomeView: View {
    @State private var categories: [Category] = []
    @EnvironmentObject private var syncManager: SyncManager

    var body: some View {
        NavigationStack {
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
            .safeAreaInset(edge: .bottom) {
                SyncStatusBar(state: syncManager.state)
            }
            .navigationTitle("羽球大师课")
            .navigationDestination(for: Category.self) { category in
                CategoryView(category: category)
            }
            .onAppear {
                categories = Database.shared.categories(parentId: nil)
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
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 6)
                    .background(.ultraThinMaterial)
                    .transition(.opacity)
            }
        }
        .animation(.easeInOut(duration: 0.3), value: state)
    }
}
