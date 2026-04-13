import SwiftUI

@main
struct YuQiuPuApp: App {
    var body: some Scene {
        WindowGroup {
            HomeView()
                .onAppear {
                    DataSync.syncIfNeeded()
                }
        }
    }
}
