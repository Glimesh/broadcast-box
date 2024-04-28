import React, { useState } from 'react';

/**
 * TreeNode component representing a single node in the expandable tree.
 * @param {Object} props - The properties passed to the component.
 * @param {string} props.name - The name of the node.
 * @param {any} props.value - The value of the node.
 * @param {Function} props.formatter - A callback function to format the node value.
 * @param {Array} props.children - An array of child nodes, which are objects of the same type as the parent node.
 */
function TreeNode({ name, value, formatter, children }) {
  const [isExpanded, setIsExpanded] = useState(false);

  /**
   * Toggle the expansion state of the tree node.
   */
  const toggleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  return (
    <div>
      <div onClick={toggleExpand} style={{ fontWeight: 'bold', cursor: 'pointer' }}>
        {name} {formatter ? formatter(value):value}
      </div>
      {isExpanded && (
        <div style={{ marginLeft: '20px' }}>
          {children?.map((child, index) => (
            <TreeNode key={index} {...child} />
          ))}
        </div>
      )}
    </div>
  );
}

export default TreeNode;
