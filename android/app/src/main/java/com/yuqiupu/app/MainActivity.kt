package com.yuqiupu.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.yuqiupu.app.ui.YuQiuPuApp
import com.yuqiupu.app.ui.YuQiuPuTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            YuQiuPuTheme {
                YuQiuPuApp()
            }
        }
    }
}
