import SwiftUI

struct PathDetailView: View {
    let path: LearningPath

    @State private var steps: [PathStep] = []
    @State private var stepContents: [Int: [ContentItem]] = [:]
    @State private var selectedURL: URL?
    @EnvironmentObject private var userState: UserState

    var body: some View {
        List {
            // Header section
            Section {
                VStack(alignment: .leading, spacing: 8) {
                    HStack {
                        DifficultyBadgeInline(difficulty: path.difficulty)
                        Spacer()
                        let completed = userState.pathProgress[path.id]?.count ?? 0
                        Text("\(completed)/\(path.stepCount) 完成")
                            .font(.caption)
                            .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                    }

                    if !path.summary.isEmpty {
                        Text(path.summary)
                            .font(.subheadline)
                            .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                    }
                }
                .listRowSeparator(.hidden)
            }

            // Steps
            ForEach(steps) { step in
                Section {
                    // Step header with day number, title, and completion toggle
                    VStack(alignment: .leading, spacing: 6) {
                        HStack(spacing: 8) {
                            // Completion toggle
                            Button {
                                userState.toggleStepComplete(pathId: path.id, stepId: step.id)
                            } label: {
                                Image(systemName: userState.isStepComplete(pathId: path.id, stepId: step.id) ? "checkmark.circle.fill" : "circle")
                                    .foregroundStyle(userState.isStepComplete(pathId: path.id, stepId: step.id) ? Color(red: 0x00/255.0, green: 0x7D/255.0, blue: 0x48/255.0) : Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                                    .font(.title3)
                            }
                            .buttonStyle(.plain)

                            if !step.day.isEmpty {
                                Text(step.day)
                                    .font(.caption2)
                                    .fontWeight(.medium)
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 2)
                                    .background(Color(red: 0x11/255.0, green: 0x11/255.0, blue: 0x11/255.0))
                                    .foregroundStyle(.white)
                                    .clipShape(Capsule())
                            }

                            Text(step.title)
                                .font(.headline)
                        }

                        if !step.note.isEmpty {
                            Text(step.note)
                                .font(.subheadline)
                                .foregroundStyle(Color(red: 0x70/255.0, green: 0x70/255.0, blue: 0x72/255.0))
                        }
                    }
                    .listRowSeparator(.hidden)

                    // Content items for this step
                    if let contents = stepContents[step.id], !contents.isEmpty {
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
            }
        }
        .navigationTitle(path.title)
        .sheet(item: $selectedURL) { url in
            SafariView(url: url)
                .ignoresSafeArea()
        }
        .task {
            steps = await Database.shared.pathStepsAsync(pathId: path.id)
            // Load contents for all steps concurrently
            await withTaskGroup(of: (Int, [ContentItem]).self) { group in
                for step in steps {
                    group.addTask {
                        let contents = await Database.shared.pathStepContentsAsync(stepId: step.id)
                        return (step.id, contents)
                    }
                }
                for await (stepId, contents) in group {
                    stepContents[stepId] = contents
                }
            }
        }
    }
}

// MARK: - Difficulty Badge (inline variant for detail view)

private struct DifficultyBadgeInline: View {
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
            .padding(.horizontal, 8)
            .padding(.vertical, 2)
            .background(badgeColor.opacity(0.15))
            .foregroundStyle(badgeColor)
            .clipShape(Capsule())
    }
}
