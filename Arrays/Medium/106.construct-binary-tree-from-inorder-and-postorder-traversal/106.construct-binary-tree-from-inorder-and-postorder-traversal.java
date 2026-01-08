/**
 * Definition for a binary tree node.
 * public class TreeNode {
 *     int val;
 *     TreeNode left;
 *     TreeNode right;
 *     TreeNode() {}
 *     TreeNode(int val) { this.val = val; }
 *     TreeNode(int val, TreeNode left, TreeNode right) {
 *         this.val = val;
 *         this.left = left;
 *         this.right = right;
 *     }
 * }
 */
class Solution {
    int index;
    Map<Integer, Integer>inorderIndexMap;
    public TreeNode buildTree(int[] inorder, int[] postorder) {
        index=postorder.length-1;
        inorderIndexMap=IntStream.range(0, inorder.length)
            .boxed().collect(Collectors.toMap(
                i->inorder[i],
                i->i,
                (existing, replacement)->existing
            ));
        return helper(inorder, postorder, 0, postorder.length-1);
    }

    TreeNode helper(int []inorder, int[]postorder, int left, int right){
        if(left>right || index<0) return null;
        TreeNode node=new TreeNode(postorder[index--]);
        int mid=inorderIndexMap.get(node.val);
        node.right=helper(inorder, postorder, mid+1, right);
        node.left=helper(inorder, postorder, left, mid-1);
        return node;
    }
}