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
    class NodeInfo{
        TreeNode node;
        int row;
        int col;
        NodeInfo(TreeNode node, int row, int col){
            this.node = node;
            this.row = row;
            this.col = col;
        }
    }   

    public List<List<Integer>> verticalTraversal(TreeNode root) {
        Map<Integer, Map<Integer, List<Integer>>> cache = new TreeMap<>();
        Deque<NodeInfo>q=new ArrayDeque<>();
        q.add(new NodeInfo(root, 0, 0));
        int level=0, ver=0;
        while(!q.isEmpty()){
            NodeInfo nodeInfo=q.poll();
            TreeNode node = nodeInfo.node;
            int row = nodeInfo.row;
            int col = nodeInfo.col;

            cache.computeIfAbsent(col, k -> new TreeMap<>())
                .computeIfAbsent(row, k -> new ArrayList<>())
                .add(node.val);

            if(node.left!=null) q.offer(new NodeInfo(node.left, row+1, col-1));
            if(node.right!=null) q.offer(new NodeInfo(node.right, row+1, col+1));
        }

        List<List<Integer>>res=new ArrayList<>();
        for(Map<Integer, List<Integer>> rows:cache.values()){
            List<Integer> column = new ArrayList<>();
            for (List<Integer> values : rows.values()) {
                Collections.sort(values);
                column.addAll(values);
            }
            res.add(column);
        }
        return res;
    }
}