import SwiftUI

@main
struct BMCApp: App {
    @StateObject private var syncManager = SyncManager.shared
    @StateObject private var userState = UserState()

    var body: some Scene {
        WindowGroup {
            HomeView()
                .environmentObject(syncManager)
                .environmentObject(userState)
                .onAppear {
                    DataSync.syncIfNeeded()
                }
        }
    }
}
