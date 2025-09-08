import React, { useState, useCallback } from 'react';

const FileUpload = ({ onUpload, acceptedTypes = ['.gpx'] }) => {
  const [isDragging, setIsDragging] = useState(false);
  const [previewContent, setPreviewContent] = useState(null);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const handleDragOver = (e) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => {
    setIsDragging(false);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
    
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      handleFile(files[0]);
    }
  };

  const handleFileInput = (e) => {
    const files = e.target.files;
    if (files.length > 0) {
      handleFile(files[0]);
    }
  };

  const handleFile = async (file) => {
    setError(null);
    
    // Validate file type
    const fileExt = file.name.split('.').pop().toLowerCase();
    if (!acceptedTypes.includes(`.${fileExt}`)) {
      setError(`Invalid file type. Supported types: ${acceptedTypes.join(', ')}`);
      return;
    }
    
    // Validate file size (max 10MB)
    if (file.size > 10 * 1024 * 1024) {
      setError('File size exceeds 10MB limit');
      return;
    }
    
    try {
      setIsLoading(true);
      
      // Preview GPX content
      if (fileExt === 'gpx') {
        const content = await file.text();
        setPreviewContent(content);
      } else {
        setPreviewContent(null);
      }
      
      // Pass file to parent component for upload
      if (onUpload) {
        onUpload(file);
      }
    } catch (err) {
      console.error('File processing error:', err);
      setError('Failed to process file');
    } finally {
      setIsLoading(false);
    }
  };

  const clearPreview = () => {
    setPreviewContent(null);
  };

  return (
    <div className="file-upload">
      <div 
        className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
          isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-blue-400'
        }`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => document.getElementById('file-input').click()}
      >
        <input
          id="file-input"
          type="file"
          className="hidden"
          accept={acceptedTypes.join(',')}
          onChange={handleFileInput}
        />
        
        <div className="flex flex-col items-center justify-center">
          {isLoading ? (
            <>
              <svg className="animate-spin h-10 w-10 text-blue-500 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <p className="mt-2 text-gray-600">Processing file...</p>
            </>
          ) : (
            <>
              <svg className="h-10 w-10 text-gray-400 mx-auto" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              <p className="mt-2 text-sm text-gray-600">
                <span className="font-medium text-blue-600">Click to upload</span> or drag and drop
              </p>
              <p className="text-xs text-gray-500 mt-1">
                {acceptedTypes.join(', ')} files, max 10MB
              </p>
            </>
          )}
        </div>
      </div>
      
      {error && (
        <div className="mt-2 p-2 bg-red-50 text-red-700 text-sm rounded-md">
          {error}
        </div>
      )}
      
      {previewContent && (
        <div className="mt-4">
          <div className="flex justify-between items-center mb-2">
            <h3 className="font-medium text-gray-800">File Preview</h3>
            <button 
              onClick={clearPreview}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              Clear preview
            </button>
          </div>
          <div className="bg-gray-50 p-3 rounded-md border border-gray-200 max-h-60 overflow-auto">
            <pre className="text-xs text-gray-700 whitespace-pre-wrap">
              {previewContent}
            </pre>
          </div>
        </div>
      )}
    </div>
  );
};

export default FileUpload;