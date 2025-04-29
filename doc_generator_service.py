import os
import base64
from io import BytesIO
from flask import Flask, request, jsonify
from docx import Document
from docx.shared import RGBColor
from bs4 import BeautifulSoup
from docx.enum.text import WD_ALIGN_PARAGRAPH

app = Flask(__name__)

# Environment variable for the save path. Defaults to /tmp if not set.
SAVE_PATH = os.environ.get("SAVE_PATH", os.getcwd())

def set_font_color(run, rgb_color):
    """Sets the font color of a run using RGB hexadecimal representation.

    Args:
        run: The docx.text.Run object to set the color for.
        rgb_color: A tuple representing the RGB color (e.g., (255, 0, 0) for red).
    """
    r, g, b = rgb_color
    hex_color = '%02x%02x%02x' % (r, g, b)  # Convert RGB tuple to hex string
    run.font.color.rgb = RGBColor.from_string(hex_color)  # Set color using hex

def html_to_docx(html_content):
    """Converts HTML content (with <article> as the root) to a docx document object.

    Parses the HTML, removes dummy spans, and populates the document with content and styles.

    Args:
        html_content: The HTML content as a string.

    Returns:
        A docx.Document object.
    """
    soup = BeautifulSoup(html_content, 'html.parser')
    document = Document()
    article = soup.find('article')

    if article:
        for element in article.children:
            if element.name == 'p':
                paragraph = document.add_paragraph()
                for child in list(element.contents):  # Iterate over a copy to allow modification
                    if child.name == 'span' and child.get('name') == 'dummy':
                        # There are dummy spans in the html_content.
                        # They are used to force the browser to render certain things inside p tags
                        # This loop replaces the "dummy" span with its children which ofter are other
                        # spans that has styling we are interested in conserving
                        for sub_child in child.contents:
                            if isinstance(sub_child, str):
                                paragraph.add_run(sub_child)
                            elif sub_child.name == 'span':
                                run = paragraph.add_run(sub_child.text)
                                if sub_child.get('style'):
                                    styles = sub_child['style'].split(';')
                                    for style in styles:
                                        style = style.strip().lower()
                                        if style == 'font-weight: bold':
                                            run.bold = True
                                        elif style == 'font-style: italic':
                                            run.italic = True
                                        elif style == 'text-decoration: underline':
                                            run.underline = True
                                        elif style.startswith('color:'):
                                            color_str = style.split(':')[1].strip()
                                            if color_str.startswith('rgb('):
                                                try:
                                                    r, g, b = map(int, color_str[4:-1].split(','))
                                                    set_font_color(run, (r, g, b))
                                                except ValueError:
                                                    print("Invalid RGB color format:", color_str)
                            elif sub_child.name == 'div':
                                for grandchild in sub_child.contents:
                                    if isinstance(grandchild, str):
                                        paragraph.add_run(grandchild)
                                    elif grandchild.name == 'span':
                                        run = paragraph.add_run(grandchild.text)
                                        if grandchild.get('style'):
                                            styles = grandchild['style'].split(';')
                                            for style in styles:
                                                style = style.strip().lower()
                                                # Apply styles to the grandchild span
                                                if style == 'font-weight: bold':
                                                    run.bold = True
                                                elif style == 'font-style: italic':
                                                    run.italic = True
                                                elif style == 'text-decoration: underline':
                                                    run.underline = True
                                                elif style.startswith('color:'):
                                                    color_str = style.split(':')[1].strip()
                                                    if color_str.startswith('rgb('):
                                                        try:
                                                            r, g, b = map(int, color_str[4:-1].split(','))
                                                            set_font_color(run, (r, g, b))
                                                        except ValueError:
                                                            print("Invalid RGB color format:", color_str)
                                    elif grandchild.name == 'button':
                                        pass # do not render button text in the document
                    elif isinstance(child, str):
                        paragraph.add_run(child)
                    elif child.name == 'span':
                        # Process regular spans (not the dummy ones)
                        run = paragraph.add_run(child.text)
                        if child.get('style'):
                            styles = child['style'].split(';')
                            for style in styles:
                                style = style.strip().lower()
                                if style == 'font-weight: bold':
                                    run.bold = True
                                elif style == 'font-style: italic':
                                    run.italic = True
                                elif style == 'text-decoration: underline':
                                    run.underline = True
                                elif style.startswith('color:'):
                                    color_str = style.split(':')[1].strip()
                                    if color_str.startswith('rgb('):
                                        try:
                                            r, g, b = map(int, color_str[4:-1].split(','))
                                            set_font_color(run, (r, g, b))
                                        except ValueError:
                                            print("Invalid RGB color format:", color_str)
                    elif child.name == 'div':
                        for grandchild in child.contents:
                            if isinstance(grandchild, str):
                                paragraph.add_run(grandchild)
                            elif grandchild.name == 'span':
                                run = paragraph.add_run(grandchild.text)
                                if grandchild.get('style'):
                                    styles = grandchild['style'].split(';')
                                    for style in styles:
                                        style = style.strip().lower()
                                        # Apply styles to spans within divs
                                        if style == 'font-weight: bold':
                                            run.bold = True
                            elif grandchild.name == 'button':
                                pass # do not render button text in the document

                # Apply paragraph styles
                if element.get('style'):
                    styles = element['style'].split(';')
                    for style in styles:
                        style = style.strip().lower()
                        if style.startswith('text-align:'):
                            align = style.split(':')[1].strip()
                            if align == 'center':
                                paragraph.alignment = WD_ALIGN_PARAGRAPH.CENTER
                            elif align == 'right':
                                paragraph.alignment = WD_ALIGN_PARAGRAPH.RIGHT
                            elif align == 'left':
                                paragraph.alignment = WD_ALIGN_PARAGRAPH.LEFT
                            elif align == 'justify':
                                paragraph.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY

    return document

@app.route('/generate_and_save', methods=['POST'])
def generate_and_save():
    """Generates a docx document from HTML content and saves it to disk.

    Receives HTML content in the request body, converts it to a docx document,
    and saves the document to the path specified by the SAVE_PATH environment variable.

    Returns:
        A JSON response with a success message and the file path, or an error message.
    """
    try:
        html_content = request.get_data(as_text=True)  # Get HTML from request body
        document = html_to_docx(html_content)  # Convert HTML to docx

        filename = "generated_document.docx"  # Generate filename
        filepath = os.path.join(SAVE_PATH, filename)  # Construct full file path

        document.save(filepath)  # Save the document
        return jsonify({"message": "Document generated and saved", "filepath": filepath}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500


@app.route('/generate_and_encode', methods=['POST'])
def generate_and_encode():
    """Generates a docx document from HTML and returns it as a base64 encoded string.

    Receives HTML content, converts it to docx, and encodes the docx file to base64.

    Returns:
        A JSON response containing the base64 encoded docx document, or an error.
    """
    try:
        html_content = request.get_data(as_text=True)  # Get HTML content
        document = html_to_docx(html_content)  # Convert HTML to docx

        buffer = BytesIO()  # Use an in-memory buffer
        document.save(buffer)  # Save docx to buffer
        buffer.seek(0)  # Reset buffer position

        base64_encoded = base64.b64encode(buffer.getvalue()).decode('utf-8')  # Encode to base64
        return jsonify({"document": base64_encoded}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 500


if __name__ == '__main__':
    from docx.enum.text import WD_ALIGN_PARAGRAPH  # Import alignment enum

    # Run the Flask app.  host='0.0.0.0' makes it accessible externally.
    app.run(debug=True, host='0.0.0.0', port=int(os.environ.get("PORT", 5000)))