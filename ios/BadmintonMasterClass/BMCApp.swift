import SwiftUI

@main
struct BMCApp: App {
    @StateObject private var syncManager = SyncManager.shared

    var body: some Scene {
        WindowGroup {
            HomeView()
                .environmentObject(syncManager)
                .onAppear {
                    DataSync.syncIfNeeded()
                }
        }
    }
}
