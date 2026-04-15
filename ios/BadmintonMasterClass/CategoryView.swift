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
        .task {
            subcategories = await Database.shared.categoriesAsync(parentId: category.id)
            contents = await Database.shared.contentsAsync(categoryId: category.id)
        }
    }
}

struct ContentRow: View {
    let item: ContentItem
    var showHeart: Bool = true
    @EnvironmentObject private var userState: UserState

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

            if showHeart {
                Spacer()

                Button {
                    userState.toggleFavorite(contentId: item.id)
                } label: {
                    Image(systemName: userState.isFavorite(contentId: item.id) ? "heart.fill" : "heart")
                        .font(.system(size: 18))
                        .foregroundStyle(userState.isFavorite(contentId: item.id)
                            ? Color(red: 0xD3/255.0, green: 0x00/255.0, blue: 0x05/255.0)
                            : Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                }
                .buttonStyle(.plain)
                .frame(width: 44, height: 44)
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
        case "bilibili": return Color(red: 0xFB/255.0, green: 0x72/255.0, blue: 0x99/255.0)   // #FB7299
        case "xiaohongshu": return Color(red: 0xFF/255.0, green: 0x24/255.0, blue: 0x42/255.0) // #FF2442
        case "douyin": return Color(red: 0x16/255.0, green: 0x18/255.0, blue: 0x23/255.0)      // #161823
        case "wechat": return Color(red: 0x07/255.0, green: 0xC1/255.0, blue: 0x60/255.0)      // #07C160
        case "youtube": return Color(red: 0xFF/255.0, green: 0x00/255.0, blue: 0x00/255.0)     // #FF0000
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
