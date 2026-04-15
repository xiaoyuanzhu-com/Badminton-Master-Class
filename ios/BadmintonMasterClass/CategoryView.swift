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
                                Spacer()
                                if sub.contentCount > 0 {
                                    Text("\(sub.contentCount) 个内容")
                                        .font(.caption)
                                        .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                                }
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
    @State private var showEditorNotes = false

    private var platformActionText: String {
        switch item.sourcePlatform {
        case "bilibili": return "在B站观看"
        case "xiaohongshu": return "在小红书查看"
        case "douyin": return "在抖音观看"
        case "wechat": return "在微信查看"
        case "youtube": return "在YouTube观看"
        default: return "打开链接"
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
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

                    // Metadata row: platform badge, author, category, difficulty, duration
                    HStack(spacing: 8) {
                        PlatformBadge(platform: item.sourcePlatform)

                        if !item.categoryName.isEmpty {
                            CategoryBadge(name: item.categoryName)
                        }

                        if !item.authorName.isEmpty {
                            Text(item.authorName)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }

                        if !item.difficulty.isEmpty {
                            ContentDifficultyBadge(difficulty: item.difficulty)
                        }

                        if !item.duration.isEmpty {
                            Text(item.duration)
                                .font(.caption)
                                .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                        }
                    }

                    // External link indicator
                    HStack(spacing: 4) {
                        Image(systemName: "arrow.up.right.square")
                            .font(.caption2)
                            .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                        Text(platformActionText)
                            .font(.caption2)
                            .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
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

            // Editor's notes (expandable)
            if !item.editorNotes.isEmpty {
                Button {
                    withAnimation(.easeInOut(duration: 0.2)) {
                        showEditorNotes.toggle()
                    }
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "note.text")
                            .font(.caption2)
                        Text("编辑笔记")
                            .font(.caption2)
                            .fontWeight(.medium)
                        Image(systemName: showEditorNotes ? "chevron.up" : "chevron.down")
                            .font(.caption2)
                    }
                    .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                }
                .buttonStyle(.plain)
                .padding(.top, 6)
                .padding(.leading, 72) // align with text stack (60 thumbnail + 12 gap)

                if showEditorNotes {
                    Text(item.editorNotes)
                        .font(.caption)
                        .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                        .padding(.top, 4)
                        .padding(.leading, 72)
                        .padding(.trailing, 16)
                        .transition(.opacity)
                }
            }
        }
        .padding(.vertical, 4)
    }
}

struct ContentDifficultyBadge: View {
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
        case "beginner": return Color(red: 0x00/255.0, green: 0x7D/255.0, blue: 0x48/255.0)
        case "intermediate": return Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0)
        case "advanced": return Color(red: 0xD3/255.0, green: 0x00/255.0, blue: 0x05/255.0)
        default: return .gray
        }
    }

    var body: some View {
        Text(displayName)
            .font(.caption2)
            .fontWeight(.medium)
            .padding(.horizontal, 6)
            .padding(.vertical, 1)
            .background(badgeColor.opacity(0.15))
            .foregroundStyle(badgeColor)
            .clipShape(Capsule())
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

struct CategoryBadge: View {
    let name: String

    var body: some View {
        Text(name)
            .font(.caption2)
            .fontWeight(.medium)
            .padding(.horizontal, 8)
            .padding(.vertical, 2)
            .background(Color(red: 0xF5/255.0, green: 0xF5/255.0, blue: 0xF5/255.0))
            .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
            .clipShape(Capsule())
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
