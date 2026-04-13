import SwiftUI

struct HomeView: View {
    @State private var categories: [Category] = []

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
