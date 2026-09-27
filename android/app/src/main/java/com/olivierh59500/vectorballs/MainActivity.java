package com.olivierh59500.vectorballs;

import android.app.Activity;
import android.os.Build;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.view.WindowInsets;
import android.view.WindowInsetsController;
import android.view.WindowManager;

import java.lang.reflect.Method;

import com.olivierh59500.vectorballsmobile.EbitenView;

import go.Seq;

public final class MainActivity extends Activity {
    private EbitenView gameView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        Seq.setContext(getApplicationContext());
        configureDCKPreview();
        // Keep the demo visible during unattended playback.
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
        gameView = new EbitenView(this);
        setContentView(gameView);
        hideSystemBars();
    }

    private void configureDCKPreview() {
        if (!getIntent().hasExtra("dck_object")) {
            return;
        }
        String name = getIntent().getStringExtra("dck_object");
        String fill = getIntent().getStringExtra("dck_fill");
        if (name == null) {
            name = "";
        }
        if (fill == null) {
            fill = "edges";
        }
        try {
            // The preserved original binding has no optional DCK preset API.
            Class<?> binding = Class.forName("com.olivierh59500.vectorballsmobile.Vectorballsmobile");
            Method configure = binding.getMethod("configurePreview", String.class, String.class,
                    long.class, double.class, long.class);
            String error = (String) configure.invoke(null, name, fill,
                    (long) getIntent().getIntExtra("dck_segments", 6),
                    getIntent().getDoubleExtra("dck_size", 640.0),
                    (long) getIntent().getIntExtra("dck_ball", -1));
            if (error != null && !error.isEmpty()) {
                Log.w("Vectorballs", "Ignoring invalid DCK preview: " + error);
            }
        } catch (ReflectiveOperationException error) {
            Log.w("Vectorballs", "DCK preview is unavailable in this build", error);
        }
    }

    private void hideSystemBars() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            WindowInsetsController controller = getWindow().getInsetsController();
            if (controller != null) {
                controller.hide(WindowInsets.Type.systemBars());
                controller.setSystemBarsBehavior(
                        WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
            }
            return;
        }

        getWindow().getDecorView().setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
                        | View.SYSTEM_UI_FLAG_FULLSCREEN
                        | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                        | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                        | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                        | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION);
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            hideSystemBars();
        }
    }

    @Override
    protected void onPause() {
        if (gameView != null) {
            gameView.suspendGame();
        }
        super.onPause();
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (gameView != null) {
            gameView.resumeGame();
        }
    }
}
