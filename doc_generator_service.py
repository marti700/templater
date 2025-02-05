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
SAVE_PATH = os.environ.get("SAVE_PATH", "/tmp")

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

    Parses the HTML, creates a docx document, and populates it with content and styles
    from the HTML.

    Args:
        html_content: The HTML content as a string.

    Returns:
        A docx.Document object.
    """

    soup = BeautifulSoup(html_content, 'html.parser')  # Parse the HTML
    document = Document()  # Create a new docx document

    article = soup.find('article')  # Find the <article> tag

    if article:
        for element in article.children: # Iterate over the children of the <article> tag
            if element.name == 'p':  # Process paragraph elements
                paragraph = document.add_paragraph()  # Add a paragraph to the document
                for span in element.find_all('span'):  # Process span elements within the paragraph
                    text = span.text  # Extract the text from the span
                    run = paragraph.add_run(text)  # Add the text as a run to the paragraph

                    # Apply styles from the span tag
                    if span.get('style'):
                        styles = span['style'].split(';')
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
                                        r, g, b = map(int, color_str[4:-1].split(',')) # Extract RGB values
                                        set_font_color(run, (r, g, b)) # Set the font color
                                    except ValueError:
                                        print("Invalid RGB color format:", color_str)
                            elif style.startswith('background-color:'):
                                pass  # TODO: Background color handling (not fully implemented, to many background color options)
                    if element.get('style'):  # Apply styles from the paragraph tag
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
                            elif style.startswith('margin-left:'):
                                pass  # Margin handling (not implemented)
                            elif style.startswith('text-indent:'):
                                pass  # Text indent handling (not implemented)

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