import SwiftUI

struct CategoryView: View {
    let category: Category

    @State private var subcategories: [Category] = []
    @State private var contents: [ContentItem] = []
    @State private var selectedURL: URL?

    var body: some View {
        List {
            if !subcategories.isEmpty {
                Section("子分类") {
                    ForEach(subcategories) { sub in
                        NavigationLink(value: sub) {
                            HStack(spacing: 12) {
                                Text(sub.icon)
                                    .font(.title3)
                                Text(sub.name)
                                    .font(.body)
                            }
                            .padding(.vertical, 2)
                        }
                    }
                }
            }

            if !contents.isEmpty {
                Section("内容") {
                    ForEach(contents) { item in
                        Button {
                            if let url = URL(string: item.sourceUrl) {
                                selectedURL = url
                            }
                        } label: {
                            ContentRow(item: item)
                        }
                        .tint(.primary)
                    }
                }
            }

            if subcategories.isEmpty && contents.isEmpty {
                ContentUnavailableView(
                    "暂无内容",
                    systemImage: "folder",
                    description: Text("该分类下还没有内容")
                )
                .listRowSeparator(.hidden)
            }
        }
        .navigationTitle(category.name)
        .navigationDestination(for: Category.self) { sub in
            CategoryView(category: sub)
        }
        .sheet(item: $selectedURL) { url in
            SafariView(url: url)
                .ignoresSafeArea()
        }
        .onAppear {
            subcategories = Database.shared.categories(parentId: category.id)
            contents = Database.shared.contents(categoryId: category.id)
        }
    }
}

struct ContentRow: View {
    let item: ContentItem

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            ContentThumbnail(url: item.thumbnailUrl)

            VStack(alignment: .leading, spacing: 6) {
                Text(item.title)
                    .font(.headline)

                if !item.summary.isEmpty {
                    Text(item.summary)
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .lineLimit(2)
                }

                HStack(spacing: 8) {
                    PlatformBadge(platform: item.sourcePlatform)

                    if !item.authorName.isEmpty {
                        Text(item.authorName)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }
            }
        }
        .padding(.vertical, 4)
    }
}

struct ContentThumbnail: View {
    let url: String

    private var imageURL: URL? {
        guard !url.isEmpty else { return nil }
        return URL(string: url)
    }

    var body: some View {
        Group {
            if let imageURL {
                AsyncImage(url: imageURL) { phase in
                    switch phase {
                    case .success(let image):
                        image
                            .resizable()
                            .aspectRatio(contentMode: .fill)
                    case .failure:
                        placeholder
                    default:
                        placeholder
                    }
                }
            } else {
                placeholder
            }
        }
        .frame(width: 60, height: 45)
        .clipShape(RoundedRectangle(cornerRadius: 6))
    }

    private var placeholder: some View {
        ZStack {
            Color(.systemGray5)
            Image(systemName: "play.rectangle.fill")
                .foregroundStyle(.secondary)
                .font(.system(size: 16))
        }
    }
}

struct PlatformBadge: View {
    let platform: String

    private var displayName: String {
        switch platform {
        case "bilibili": return "B站"
        case "xiaohongshu": return "小红书"
        case "douyin": return "抖音"
        case "wechat": return "微信"
        case "youtube": return "YouTube"
        default: return "其他"
        }
    }

    private var badgeColor: Color {
        switch platform {
        case "bilibili": return .pink
        case "xiaohongshu": return .red
        case "douyin": return .black
        case "wechat": return .green
        case "youtube": return .red
        default: return .gray
        }
    }

    var body: some View {
        Text(displayName)
            .font(.caption2)
            .fontWeight(.medium)
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .background(badgeColor.opacity(0.15))
            .foregroundStyle(badgeColor)
            .clipShape(Capsule())
    }
}
