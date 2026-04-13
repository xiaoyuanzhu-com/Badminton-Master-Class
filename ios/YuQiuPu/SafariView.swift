import SwiftUI
import SafariServices

struct SafariView: UIViewControllerRepresentable {
    let url: URL

    func makeUIViewController(context: Context) -> SFSafariViewController {
        SFSafariViewController(url: url)
    }

    func updateUIViewController(_ uiViewController: SFSafariViewController, context: Context) {}
}

// Make URL conform to Identifiable so it can be used with .sheet(item:)
extension URL: @retroactive Identifiable {
    public var id: String { absoluteString }
}
