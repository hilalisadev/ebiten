// Copyright 2019 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ebitenmobilegobind

// gobind is a wrapper of the original gobind. This command adds extra files like a view controller.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	lang          = flag.String("lang", "", "")
	outdir        = flag.String("outdir", "", "")
	javaPkg       = flag.String("javapkg", "", "")
	prefix        = flag.String("prefix", "", "")
	bootclasspath = flag.String("bootclasspath", "", "")
	classpath     = flag.String("classpath", "", "")
	tags          = flag.String("tags", "", "")
)

var usage = `The Gobind tool generates Java language bindings for Go.

For usage details, see doc.go.`

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cmd := exec.Command("gobind-original", os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	pkgs, err := packages.Load(nil, flag.Args()[0])
	if err != nil {
		return err
	}
	prefixLower := *prefix + pkgs[0].Name
	prefixUpper := strings.Title(*prefix) + strings.Title(pkgs[0].Name)

	writeFile := func(filename string, content string) error {
		if err := ioutil.WriteFile(filepath.Join(*outdir, filename), []byte(content), 0644); err != nil {
			return err
		}
		return nil
	}
	replacePrefixes := func(content string) string {
		content = strings.ReplaceAll(content, "{{.PrefixUpper}}", prefixUpper)
		content = strings.ReplaceAll(content, "{{.PrefixLower}}", prefixLower)
		content = strings.ReplaceAll(content, "{{.JavaPkg}}", *javaPkg)
		return content
	}

	// Add additional files.
	langs := strings.Split(*lang, ",")
	for _, lang := range langs {
		switch lang {
		case "objc":
			// iOS
			if err := writeFile(filepath.Join("src", "gobind", prefixLower+"ebitenviewcontroller_ios.m"), replacePrefixes(objcM)); err != nil {
				return err
			}
		case "java":
			// Android
			dir := filepath.Join(strings.Split(*javaPkg, ".")...)
			dir = filepath.Join(dir, prefixLower)
			if err := writeFile(filepath.Join("java", dir, "EbitenView.java"), replacePrefixes(viewJava)); err != nil {
				return err
			}
			if err := writeFile(filepath.Join("java", dir, "EbitenSurfaceView.java"), replacePrefixes(surfaceViewJava)); err != nil {
				return err
			}
		case "go":
			// Do nothing.
		default:
			panic(fmt.Sprintf("unsupported language: %s", lang))
		}
	}

	return nil
}

const objcM = `// Code generated by ebitenmobile. DO NOT EDIT.

// +build ios

#import <stdint.h>
#import <UIKit/UIKit.h>
#import <GLKit/GLkit.h>
#import "Ebitenmobileview.objc.h"

@interface {{.PrefixUpper}}EbitenViewController : UIViewController
@end

@implementation {{.PrefixUpper}}EbitenViewController {
  GLKView* glkView_;
}

- (GLKView*)glkView {
  if (!glkView_) {
    glkView_ = [[GLKView alloc] init];
    glkView_.multipleTouchEnabled = YES;
  }
  return glkView_;
}

- (void)viewDidLoad {
  [super viewDidLoad];

  self.glkView.delegate = (id<GLKViewDelegate>)(self);
  [self.view addSubview: self.glkView];

  EAGLContext *context = [[EAGLContext alloc] initWithAPI:kEAGLRenderingAPIOpenGLES2];
  [self glkView].context = context;
	
  [EAGLContext setCurrentContext:context];
	
  CADisplayLink *displayLink = [CADisplayLink displayLinkWithTarget:self selector:@selector(drawFrame)];
  [displayLink addToRunLoop:[NSRunLoop currentRunLoop] forMode:NSDefaultRunLoopMode];
}

- (void)viewDidLayoutSubviews {
  [super viewDidLayoutSubviews];
  CGRect viewRect = [[self view] frame];

  EbitenmobileviewLayout(viewRect.size.width, viewRect.size.height, (id<EbitenmobileviewViewRectSetter>)self);
}

- (void)setViewRect:(long)x y:(long)y width:(long)width height:(long)height {
  CGRect glkViewRect = CGRectMake(x, y, width, height);
  [[self glkView] setFrame:glkViewRect];
}

- (void)didReceiveMemoryWarning {
  [super didReceiveMemoryWarning];
  // Dispose of any resources that can be recreated.
  // TODO: Notify this to Go world?
}

- (void)drawFrame{
  [[self glkView] setNeedsDisplay];
}

- (void)glkView:(GLKView*)view drawInRect:(CGRect)rect {
  NSError* err = nil;
  EbitenmobileviewUpdate(&err);
  if (err != nil) {
    NSLog(@"Error: %@", err);
  }
}

- (void)updateTouches:(NSSet*)touches {
  for (UITouch* touch in touches) {
    if (touch.view != [self glkView]) {
      continue;
    }
    CGPoint location = [touch locationInView:touch.view];
    EbitenmobileviewUpdateTouchesOnIOS(touch.phase, (uintptr_t)touch, location.x, location.y);
  }
}

- (void)touchesBegan:(NSSet*)touches withEvent:(UIEvent*)event {
  [self updateTouches:touches];
}

- (void)touchesMoved:(NSSet*)touches withEvent:(UIEvent*)event {
  [self updateTouches:touches];
}

- (void)touchesEnded:(NSSet*)touches withEvent:(UIEvent*)event {
  [self updateTouches:touches];
}

- (void)touchesCancelled:(NSSet*)touches withEvent:(UIEvent*)event {
  [self updateTouches:touches];
}

@end
`

const viewJava = `// Code generated by ebitenmobile. DO NOT EDIT.

package {{.JavaPkg}}.{{.PrefixLower}};

import android.content.Context;
import android.util.AttributeSet;
import android.util.Log;
import android.view.ViewGroup;

import {{.JavaPkg}}.ebitenmobileview.Ebitenmobileview;
import {{.JavaPkg}}.ebitenmobileview.ViewRectSetter;

public class EbitenView extends ViewGroup {
    private double getDeviceScale() {
        if (deviceScale_ == 0.0) {
            deviceScale_ = getResources().getDisplayMetrics().density;
        }
        return deviceScale_;
    }

    private double pxToDp(double x) {
        return x / getDeviceScale();
    }

    private double dpToPx(double x) {
        return x * getDeviceScale();
    }

    private double deviceScale_ = 0.0;

    public EbitenView(Context context) {
        super(context);
        ebitenSurfaceView_ = new EbitenSurfaceView(context);
    }

    public EbitenView(Context context, AttributeSet attrs) {
        super(context, attrs);
        ebitenSurfaceView_ = new EbitenSurfaceView(context, attrs);
    }

    @Override
    protected void onLayout(boolean changed, int left, int top, int right, int bottom) {
        if (!initialized_) {
            LayoutParams params = new LayoutParams(LayoutParams.WRAP_CONTENT, LayoutParams.WRAP_CONTENT);
            addView(ebitenSurfaceView_, params);
            initialized_ = true;
        }

        int widthInDp = (int)Math.floor(pxToDp(right - left));
        int heightInDp = (int)Math.floor(pxToDp(bottom - top));
        Ebitenmobileview.layout(widthInDp, heightInDp, new ViewRectSetter() {
            @Override
            public void setViewRect(long xInDp, long yInDp, long widthInDp, long heightInDp) {
                int widthInPx = (int)Math.ceil(dpToPx(widthInDp));
                int heightInPx = (int)Math.ceil(dpToPx(heightInDp));
                int xInPx = (int)Math.ceil(dpToPx(xInDp));
                int yInPx = (int)Math.ceil(dpToPx(yInDp));
                ebitenSurfaceView_.layout(xInPx, yInPx, xInPx + widthInPx, yInPx + heightInPx);
            }
        });
    }

    public void onPause() {
        if (initialized_) {
            ebitenSurfaceView_.onPause();
        }
    }

    public void onResume() {
        if (initialized_) {
            ebitenSurfaceView_.onPause();
        }
    }

    protected void onErrorOnGameUpdate(Exception e) {
        Log.e("Go", e.toString());
    }

    private EbitenSurfaceView ebitenSurfaceView_;
    private boolean initialized_ = false;
}
`

const surfaceViewJava = `// Code generated by ebitenmobile. DO NOT EDIT.

package {{.JavaPkg}}.{{.PrefixLower}};

import android.content.Context;
import android.opengl.GLSurfaceView;
import android.os.Handler;
import android.os.Looper;
import android.util.AttributeSet;
import android.view.MotionEvent;

import javax.microedition.khronos.egl.EGLConfig;
import javax.microedition.khronos.opengles.GL10;

import {{.JavaPkg}}.ebitenmobileview.Ebitenmobileview;
import {{.JavaPkg}}.{{.PrefixLower}}.EbitenView;

class EbitenSurfaceView extends GLSurfaceView {

    private class EbitenRenderer implements GLSurfaceView.Renderer {

        private boolean errored_ = false;

        @Override
        public void onDrawFrame(GL10 gl) {
            if (errored_) {
                return;
            }
            try {
                Ebitenmobileview.update();
            } catch (final Exception e) {
                new Handler(Looper.getMainLooper()).post(new Runnable() {
                    @Override
                    public void run() {
                        onErrorOnGameUpdate(e);
                    }
                });
                errored_ = true;
            }
        }

        @Override
        public void onSurfaceCreated(GL10 gl, EGLConfig config) {
        }

        @Override
        public void onSurfaceChanged(GL10 gl, int width, int height) {
        }
    }

    public EbitenSurfaceView(Context context) {
        super(context);
        initialize();
    }

    public EbitenSurfaceView(Context context, AttributeSet attrs) {
        super(context, attrs);
        initialize();
    }

    private void initialize() {
        setEGLContextClientVersion(2);
        setEGLConfigChooser(8, 8, 8, 8, 0, 0);
        setRenderer(new EbitenRenderer());
    }

    private double getDeviceScale() {
        if (deviceScale_ == 0.0) {
            deviceScale_ = getResources().getDisplayMetrics().density;
        }
        return deviceScale_;
    }

    private double pxToDp(double x) {
        return x / getDeviceScale();
    }

    @Override
    public boolean onTouchEvent(MotionEvent e) {
        for (int i = 0; i < e.getPointerCount(); i++) {
            int id = e.getPointerId(i);
            int x = (int)e.getX(i);
            int y = (int)e.getY(i);
            Ebitenmobileview.updateTouchesOnAndroid(e.getActionMasked(), id, (int)pxToDp(x), (int)pxToDp(y));
        }
        return true;
    }

    private void onErrorOnGameUpdate(Exception e) {
        ((EbitenView)getParent()).onErrorOnGameUpdate(e);
    }

    private double deviceScale_ = 0.0;
}
`