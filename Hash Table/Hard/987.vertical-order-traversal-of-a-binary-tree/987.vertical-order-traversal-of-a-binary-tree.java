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
    public void mapping(TreeMap<Integer, TreeMap<Integer, PriorityQueue<Integer>>> map, TreeNode root, int x, int l){
        if(root == null){
            return;
        }
        if(!map.containsKey(x)){
            map.put(x, new TreeMap());
        }

        if(!map.get(x).containsKey(l)){
            map.get(x).put(l, new PriorityQueue<>());
        }

        map.get(x).get(l).add(root.val);

        mapping(map, root.left, x-1, l+1);
        mapping(map, root.right, x+1, l+1);
    }

    public List<List<Integer>> verticalTraversal(TreeNode root) {
        if(root == null){
            return new LinkedList<>();
        }
        List<List<Integer>> ans = new LinkedList<>();
        
        TreeMap<Integer, TreeMap<Integer, PriorityQueue<Integer>>> map = new TreeMap<>();
        mapping(map, root, 0, 0);

        for(Map.Entry<Integer, TreeMap<Integer, PriorityQueue<Integer>>> entry : map.entrySet()){
            List<Integer> l = new LinkedList<>();
            for(Map.Entry<Integer, PriorityQueue<Integer>> e : entry.getValue().entrySet()){
                PriorityQueue<Integer> q = e.getValue();
                while(!q.isEmpty()){
                    l.add(q.poll());
                }
                
            }
            ans.add(l);
        }
        return ans;
    }
}