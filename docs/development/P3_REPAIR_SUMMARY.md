# Version 25 N1 P3 repair summary

Preview 2 testing exposed text clipping and an unresponsive Windows interface.

Root causes identified:

- fonts scaled for display DPI while many rectangles remained fixed-size;
- background work could interact with UI controls unsafely;
- painting could require an expensive full workspace snapshot;
- several layouts assumed one window width;
- ampersands in labels were interpreted as accelerator markers.

Preview 3 changed DPI handling, reflowed the Home layout, removed full encrypted workspace cloning from paint, routed completion through the Windows message queue, moved more processing off the interface thread and corrected label rendering.

The build remains unapproved pending direct Windows execution, accessibility and stability evidence.
