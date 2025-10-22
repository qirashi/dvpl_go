# dvpl_go [BACK](./../README_EN.md)

> [!NOTE]
> Instructions for using and installing the **dvpl** utility.

---

## Description

The **dvpl_go** utility allows you to work with files in the DVPL format. It can be used in two modes:

1. **Full Installation**: After installation, the utility integrates into the operating system's context menu, enabling quick access for selected files or folders (see screenshot below).
2. **Manual Launch**: The converter can be placed in any folder and run directly from there. In this case, two modes are available:
   - Compressing files into the DVPL format.
   - Decompressing files from the DVPL format.

---

## Installation

### Full Installation
1. Run the installer from `Releases`.
2. Follow the installer's instructions.
3. Choose the option to integrate the utility into the context menu (recommended for ease of use).

Once the installation is complete, you can use the utility via the context menu, as shown in the screenshot.

### Manual Installation
1. Download the archive for your architecture from the **Releases** section.
2. Extract the contents of the archive into any convenient folder (e.g., the desktop).
3. Use the utility directly from this folder.

---

## Usage

### Via Context Menu
1. Select the file or folder you want to process.
2. Right-click and choose the appropriate menu option (compress or decompress).
3. The utility will automatically perform the required actions.

![Context menu example](screenshot.png)  
_Example of utility integration in the context menu_

### Via Manual Launch
1. Place the `dvpl.exe` file in the target folder with the files you want to process.
2. Launch the utility in one of the following ways:
   - For **compression**: Drag and drop a file or folder onto the `dvpl.exe` executable.
   - For **decompression**: Run the utility from the command line with the appropriate parameters (see the documentation inside the archive for details).

![Example of direct program launch](screenshot_2.png)  
_Example of utility integration in the context menu_

---

## Notes
- If you use the utility without installation, ensure it is located in the same folder as the target files or specify the correct path to them.
- For additional information on command-line parameters, refer to the documentation inside the `dvpl_go.zip` archive.