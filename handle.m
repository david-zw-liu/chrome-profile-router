#include "handler.h"

@implementation BrowseAppDelegate
- (void)applicationWillFinishLaunching:(NSNotification *)aNotification
{
  NSAppleEventManager *appleEventManager = [NSAppleEventManager sharedAppleEventManager];
  [appleEventManager setEventHandler:self
                         andSelector:@selector(handleGetURLEvent:withReplyEvent:)
                         forEventClass:kInternetEventClass andEventID:kAEGetURL];
}

- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender
{
  return NSTerminateNow;
}

- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
           withReplyEvent:(NSAppleEventDescriptor *)replyEvent {
  HandleURL((char*)[[[event paramDescriptorForKeyword:keyDirectObject] stringValue] UTF8String]);
}

- (BOOL)application:(NSApplication *)sender openFile:(NSString *)filename {
  HandleURL((char*)[filename UTF8String]);
  return YES;
}
@end

void Run(void) {
  [NSAutoreleasePool new];
  [NSApplication sharedApplication];
  BrowseAppDelegate *app = [BrowseAppDelegate alloc];
  [NSApp setDelegate:app];
  [NSApp run];
}
