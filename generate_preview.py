from PIL import Image, ImageDraw, ImageFont
import os

# Create image
width, height = 1280, 640
img = Image.new('RGB', (width, height), '#0d1117')
draw = ImageDraw.Draw(img)

# Try to load fonts, fall back to default
font_paths = [
    '/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf',
    '/usr/share/fonts/TTF/DejaVuSans-Bold.ttf',
    '/usr/share/fonts/dejavu/DejaVuSans-Bold.ttf',
]

def load_font(size, bold=True):
    for path in font_paths:
        if os.path.exists(path):
            return ImageFont.truetype(path, size)
    return ImageFont.load_default()

# Draw decorative circles
draw.ellipse([50, 450, 350, 750], fill='#161b22')
draw.ellipse([980, 50, 1180, 250], fill='#0969da')

# Title
title_font = load_font(80)
title_text = "dir2opds"
title_bbox = draw.textbbox((0, 0), title_text, font=title_font)
title_w = title_bbox[2] - title_bbox[0]
title_x = (width - title_w) // 2
draw.text((title_x, 240), title_text, font=title_font, fill='#ffffff')

# Tagline
tagline_font = load_font(36, bold=False)
tagline_text = "Self-Hosted OPDS Ebook Server"
tagline_bbox = draw.textbbox((0, 0), tagline_text, font=tagline_font)
tagline_w = tagline_bbox[2] - tagline_bbox[0]
tagline_x = (width - tagline_w) // 2
draw.text((tagline_x, 350), tagline_text, font=tagline_font, fill='#58a6ff')

# Description
desc_font = load_font(26, bold=False)
desc_text = "Turn any folder into a digital library"
desc_bbox = draw.textbbox((0, 0), desc_text, font=desc_font)
desc_w = desc_bbox[2] - desc_bbox[0]
desc_x = (width - desc_w) // 2
draw.text((desc_x, 410), desc_text, font=desc_font, fill='#8b949e')

# Accent line
line_y = 460
line_w = 200
line_x = (width - line_w) // 2
draw.rounded_rectangle([line_x, line_y, line_x + line_w, line_y + 4], radius=2, fill='#0969da')

# Save
output_path = '/home/dubyte/Documents/Workspace/go/dir2opds/docs/assets/social-preview.png'
img.save(output_path, 'PNG')
print(f"Saved to {output_path}")
