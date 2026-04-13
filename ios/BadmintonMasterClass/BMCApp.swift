import SwiftUI

@main
struct BMCApp: App {
    var body: some Scene {
        WindowGroup {
            HomeView()
                .onAppear {
                    DataSync.syncIfNeeded()
                }
        }
    }
}
